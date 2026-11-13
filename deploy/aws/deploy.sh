#!/usr/bin/env bash
#
# WeBond - deploy to AWS (ECS Fargate + RDS MySQL + S3/CloudFront).
#
# SAFE BY DEFAULT: without --apply this only prints the plan and changes nothing.
#
#   deploy/aws/deploy.sh --env dev                          # show the plan
#   deploy/aws/deploy.sh --env dev --apply                  # deploy (asks to confirm)
#   deploy/aws/deploy.sh --env dev --apply --with-demo-data # also load the demo dataset
#   deploy/aws/deploy.sh --env prod --params deploy/aws/params.prod.env --apply
#
# Options
#   --env dev|prod      environment (default: dev)
#   --region REGION     AWS region (default: $AWS_REGION, $AWS_DEFAULT_REGION, ap-southeast-1)
#   --profile NAME      AWS CLI profile
#   --params FILE       KEY=VALUE overrides, see deploy/aws/params.example.env
#                       (default: deploy/aws/params.<env>.env when it exists)
#   --with-demo-data    load the demo dataset (publicly documented passwords!). dev only.
#   --apply             really deploy
#   --yes               do not ask for confirmation (CI)
#   -h, --help          this help
#
# Order (so new code never runs on an old schema):
#   1 infrastructure stack   2 build + push API image   3 database migration with
#   the new image   4 roll out the new API version   5 build + upload the web app
#
# Needs: aws CLI v2, docker, node + npm, and credentials allowed to create the
# resources in template.yaml. See docs/DEPLOYMENT.md.
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE="$ROOT/deploy/aws/template.yaml"

ENVIRONMENT="dev"
REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-ap-southeast-1}}"
PROFILE=""
PARAMS_FILE=""
WITH_DEMO=0
APPLY=0
ASSUME_YES=0

usage() { sed -n '2,29p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; }
die() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }
step() { printf '\n==> %s\n' "$*"; }
note() { printf '    %s\n' "$*"; }

while [ $# -gt 0 ]; do
  case "$1" in
    --env) ENVIRONMENT="${2:-}"; shift 2 ;;
    --region) REGION="${2:-}"; shift 2 ;;
    --profile) PROFILE="${2:-}"; shift 2 ;;
    --params) PARAMS_FILE="${2:-}"; shift 2 ;;
    --with-demo-data) WITH_DEMO=1; shift ;;
    --apply) APPLY=1; shift ;;
    --yes) ASSUME_YES=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1 (try --help)" ;;
  esac
done

case "$ENVIRONMENT" in dev|prod) ;; *) die "--env must be dev or prod" ;; esac
[ -n "$REGION" ] || die "no region: pass --region or set AWS_REGION"
if [ "$WITH_DEMO" -eq 1 ] && [ "$ENVIRONMENT" = "prod" ]; then
  die "--with-demo-data is refused for prod: the demo accounts have publicly documented passwords."
fi

# Regions whose load-balancer log-delivery account is known to template.yaml.
case "$REGION" in
  ap-southeast-1|ap-southeast-2|ap-northeast-1|us-east-1|us-west-2|eu-west-1) ;;
  *) die "region $REGION is not supported by template.yaml (see its ElbLogDeliveryAccount mapping)" ;;
esac

STACK="webond-$ENVIRONMENT"
FAMILY="webond-$ENVIRONMENT-api"
DRY=$((1 - APPLY))

AWS=(aws --region "$REGION")
[ -z "$PROFILE" ] || AWS+=(--profile "$PROFILE")

# ---- helpers ----------------------------------------------------------------
# run: execute a command that changes something (or just print it in a dry run).
run() {
  if [ "$DRY" -eq 1 ]; then
    printf '    [dry-run] %s\n' "$*"
  else
    "$@"
  fi
}
# query: read-only AWS call; placeholder text in a dry run.
query() {
  local placeholder="$1"; shift
  if [ "$DRY" -eq 1 ]; then echo "<$placeholder>"; else "$@"; fi
}
stack_output() {
  query "$1" "${AWS[@]}" cloudformation describe-stacks --stack-name "$STACK" \
    --query "Stacks[0].Outputs[?OutputKey=='$1'].OutputValue" --output text
}
stack_exists() {
  [ "$DRY" -eq 0 ] && "${AWS[@]}" cloudformation describe-stacks --stack-name "$STACK" >/dev/null 2>&1
}
param_of_stack() { # current value of a stack parameter, empty if none
  "${AWS[@]}" cloudformation describe-stacks --stack-name "$STACK" \
    --query "Stacks[0].Parameters[?ParameterKey=='$1'].ParameterValue" --output text 2>/dev/null || true
}

# ---- parameters ------------------------------------------------------------
if [ -z "$PARAMS_FILE" ] && [ -f "$ROOT/deploy/aws/params.$ENVIRONMENT.env" ]; then
  PARAMS_FILE="$ROOT/deploy/aws/params.$ENVIRONMENT.env"
fi

CFN_PARAMS=()      # Key=Value pairs passed to the template
SERVICE_COUNT=""   # how many API tasks to run after the rollout
if [ -n "$PARAMS_FILE" ]; then
  [ -f "$PARAMS_FILE" ] || die "params file not found: $PARAMS_FILE"
  while IFS='=' read -r key value; do
    case "$key" in ''|\#*) continue ;; esac
    value="${value%\"}"; value="${value#\"}"
    case "$key" in
      ServiceCount) SERVICE_COUNT="$value" ;;
      DomainName|HostedZoneId|AlbCertificateArn|CloudFrontCertificateArn|AlertEmail|\
DbInstanceClass|TaskCpu|TaskMemory|LogRetentionDays|CloudFrontPriceClass|MaxCount|VpcCidr|\
SmtpHost|SmtpPort|SmtpUsername|SmtpFrom|SmtpPasswordParameter)
        CFN_PARAMS+=("$key=$value") ;;
      *) die "unknown key '$key' in $PARAMS_FILE (see deploy/aws/params.example.env)" ;;
    esac
  done <"$PARAMS_FILE"
fi
if [ -z "$SERVICE_COUNT" ]; then
  if [ "$ENVIRONMENT" = "prod" ]; then SERVICE_COUNT=2; else SERVICE_COUNT=1; fi
fi
case "$SERVICE_COUNT" in ''|*[!0-9]*) die "ServiceCount must be a number" ;; esac

# ---- image tag (tags are immutable in ECR) ---------------------------------
if SHA="$(git -C "$ROOT" rev-parse --short=12 HEAD 2>/dev/null)"; then
  IMAGE_TAG="$SHA"
  if [ -n "$(git -C "$ROOT" status --porcelain 2>/dev/null)" ]; then
    IMAGE_TAG="$SHA-dirty-$(date +%Y%m%d%H%M%S)"
  fi
else
  IMAGE_TAG="build-$(date +%Y%m%d%H%M%S)"
fi

# ---- plan -------------------------------------------------------------------
cat <<PLAN

WeBond AWS deployment
  environment : $ENVIRONMENT
  region      : $REGION
  stack       : $STACK
  image tag   : $IMAGE_TAG
  api tasks   : $SERVICE_COUNT
  params file : ${PARAMS_FILE:-none (dev defaults: *.cloudfront.net address, no custom domain)}
  demo data   : $([ "$WITH_DEMO" -eq 1 ] && echo yes || echo no)
  mode        : $([ "$APPLY" -eq 1 ] && echo "APPLY (changes AWS)" || echo "DRY RUN (changes nothing)")

PLAN

step "Preflight"
for tool in aws docker npm node; do
  command -v "$tool" >/dev/null 2>&1 || die "'$tool' is not installed"
done
echo "tools OK"
if [ "$DRY" -eq 0 ]; then
  docker info >/dev/null 2>&1 || die "the Docker daemon is not running"
  IDENTITY="$("${AWS[@]}" sts get-caller-identity --query '[Account,Arn]' --output text)" \
    || die "AWS credentials are not configured (aws configure / --profile / SSO login)"
  echo "AWS identity: $IDENTITY"
else
  note "[dry-run] skipping the AWS credential and Docker daemon checks"
fi

if [ "$APPLY" -eq 1 ] && [ "$ASSUME_YES" -eq 0 ]; then
  printf '\nThis will create/update real AWS resources that cost money.\nType the stack name (%s) to continue: ' "$STACK"
  read -r answer
  [ "$answer" = "$STACK" ] || die "not confirmed; nothing was changed"
fi

# ---- 1. infrastructure ------------------------------------------------------
step "1/5 Infrastructure stack ($STACK)"
if command -v cfn-lint >/dev/null 2>&1; then
  cfn-lint "$TEMPLATE" && echo "cfn-lint OK"
fi

PREFIX_LIST="$(query cloudfront-prefix-list "${AWS[@]}" ec2 describe-managed-prefix-lists \
  --filters Name=prefix-list-name,Values=com.amazonaws.global.cloudfront.origin-facing \
  --query 'PrefixLists[0].PrefixListId' --output text)"
note "CloudFront prefix list: $PREFIX_LIST"

# Keep the running image/count on an update; start empty (0 tasks) on the first deploy.
CUR_TAG="bootstrap"; CUR_COUNT="0"
if stack_exists; then
  CUR_TAG="$(param_of_stack ImageTag)"; CUR_COUNT="$(param_of_stack DesiredCount)"
  note "existing stack: keeping image '$CUR_TAG' with $CUR_COUNT task(s) until the rollout step"
fi

deploy_stack() { # $1 = image tag, $2 = desired count
  run "${AWS[@]}" cloudformation deploy \
    --stack-name "$STACK" \
    --template-file "$TEMPLATE" \
    --capabilities CAPABILITY_IAM \
    --no-fail-on-empty-changeset \
    --tags Project=webond Environment="$ENVIRONMENT" DataClassification=Confidential ManagedBy=cloudformation \
    --parameter-overrides \
      Environment="$ENVIRONMENT" \
      CloudFrontPrefixListId="$PREFIX_LIST" \
      ImageTag="$1" DesiredCount="$2" \
      ${CFN_PARAMS[@]+"${CFN_PARAMS[@]}"}
}
deploy_stack "$CUR_TAG" "$CUR_COUNT"

REPO_URI="$(stack_output RepositoryUri)"
CLUSTER="$(stack_output ClusterName)"
SERVICE="$(stack_output ServiceName)"
SUBNETS="$(stack_output AppSubnetIds)"
API_SG="$(stack_output ApiSecurityGroupId)"
BUCKET="$(stack_output WebBucketName)"
DIST_ID="$(stack_output DistributionId)"
LOG_GROUP="$(stack_output ApiLogGroupName)"

# ---- 2. image ---------------------------------------------------------------
step "2/5 Build and push the API image ($REPO_URI:$IMAGE_TAG)"
REGISTRY="${REPO_URI%%/*}"
if [ "$DRY" -eq 1 ]; then
  note "[dry-run] aws ecr get-login-password | docker login --username AWS --password-stdin $REGISTRY"
else
  "${AWS[@]}" ecr get-login-password | docker login --username AWS --password-stdin "$REGISTRY"
fi
run docker build --platform linux/arm64 -t "$REPO_URI:$IMAGE_TAG" "$ROOT/api"
run docker push "$REPO_URI:$IMAGE_TAG"

# ---- 3. migration -----------------------------------------------------------
# A separate task-definition family is registered with the NEW image so the
# migration can run before the service is rolled out.
run_one_off() { # $1 = command (migrate|seed), $2... = extra env overrides (NAME=VALUE)
  local cmd="$1"; shift
  local env_json=""
  local pair
  for pair in "$@"; do
    env_json="${env_json}{\"name\":\"${pair%%=*}\",\"value\":\"${pair#*=}\"},"
  done
  env_json="[${env_json%,}]"

  if [ "$DRY" -eq 1 ]; then
    note "[dry-run] register $FAMILY-oneoff with image $REPO_URI:$IMAGE_TAG"
    note "[dry-run] aws ecs run-task ... command=[$cmd] env=$env_json  (then wait, check exit code 0)"
    return 0
  fi

  local tmp current new
  tmp="$(mktemp)"
  current="$("${AWS[@]}" ecs describe-task-definition --task-definition "$FAMILY" \
    --query 'taskDefinition.containerDefinitions[0].image' --output text)"
  new="$REPO_URI:$IMAGE_TAG"
  "${AWS[@]}" ecs describe-task-definition --task-definition "$FAMILY" --output json \
    --query 'taskDefinition.{family:family,taskRoleArn:taskRoleArn,executionRoleArn:executionRoleArn,networkMode:networkMode,containerDefinitions:containerDefinitions,requiresCompatibilities:requiresCompatibilities,cpu:cpu,memory:memory,runtimePlatform:runtimePlatform}' \
    | sed -e "s#\"family\": \"$FAMILY\"#\"family\": \"$FAMILY-oneoff\"#" -e "s#$current#$new#" >"$tmp"
  "${AWS[@]}" ecs register-task-definition --cli-input-json "file://$tmp" >/dev/null
  rm -f "$tmp"

  local overrides task_arn code
  overrides="{\"containerOverrides\":[{\"name\":\"api\",\"command\":[\"$cmd\"],\"environment\":$env_json}]}"
  task_arn="$("${AWS[@]}" ecs run-task \
    --cluster "$CLUSTER" --launch-type FARGATE --task-definition "$FAMILY-oneoff" \
    --network-configuration "awsvpcConfiguration={subnets=[$SUBNETS],securityGroups=[$API_SG],assignPublicIp=DISABLED}" \
    --overrides "$overrides" --query 'tasks[0].taskArn' --output text)"
  [ -n "$task_arn" ] && [ "$task_arn" != "None" ] || die "could not start the '$cmd' task"
  note "started $cmd task: $task_arn"
  "${AWS[@]}" ecs wait tasks-stopped --cluster "$CLUSTER" --tasks "$task_arn"
  code="$("${AWS[@]}" ecs describe-tasks --cluster "$CLUSTER" --tasks "$task_arn" \
    --query 'tasks[0].containers[0].exitCode' --output text)"
  if [ "$code" != "0" ]; then
    die "the '$cmd' task failed (exit code $code). Nothing was rolled out. Logs: aws logs tail $LOG_GROUP --since 30m"
  fi
  echo "'$cmd' finished successfully"
}

step "3/5 Database migration (new image, before the rollout)"
run_one_off migrate
if [ "$WITH_DEMO" -eq 1 ]; then
  step "3b Demo data (dev only)"
  run_one_off seed WEBOND_ALLOW_PROD_SEED=1
fi

# ---- 4. rollout -------------------------------------------------------------
step "4/5 Roll out API $IMAGE_TAG with $SERVICE_COUNT task(s)"
deploy_stack "$IMAGE_TAG" "$SERVICE_COUNT"
if [ "$SERVICE_COUNT" -gt 0 ]; then
  run "${AWS[@]}" ecs wait services-stable --cluster "$CLUSTER" --services "$SERVICE"
fi

# ---- 5. web -----------------------------------------------------------------
step "5/5 Build and upload the web app"
SHOW_HINT=false; [ "$WITH_DEMO" -eq 1 ] && SHOW_HINT=true
# /api is same-origin: CloudFront routes it to the load balancer.
if [ "$DRY" -eq 1 ]; then
  note "[dry-run] (cd web && npm ci && VITE_DATA_SOURCE=api VITE_API_URL=/api VITE_SHOW_DEMO_HINT=$SHOW_HINT npm run build)"
else
  ( cd "$ROOT/web" && npm ci && \
    VITE_DATA_SOURCE=api VITE_API_URL=/api VITE_SHOW_DEMO_HINT="$SHOW_HINT" npm run build )
fi
# Hashed assets are cached for a year; everything else revalidates quickly.
run "${AWS[@]}" s3 sync "$ROOT/web/dist" "s3://$BUCKET/" --exclude "*" --include "assets/*" \
  --cache-control "public,max-age=31536000,immutable"
run "${AWS[@]}" s3 sync "$ROOT/web/dist" "s3://$BUCKET/" --exclude "assets/*" --exclude "index.html" --delete \
  --cache-control "public,max-age=300"
run "${AWS[@]}" s3 cp "$ROOT/web/dist/index.html" "s3://$BUCKET/index.html" \
  --cache-control "no-cache" --content-type "text/html; charset=utf-8"
run "${AWS[@]}" cloudfront create-invalidation --distribution-id "$DIST_ID" --paths "/index.html" "/"

# ---- done -------------------------------------------------------------------
SITE="$(stack_output SiteUrl)"
cat <<DONE

$([ "$APPLY" -eq 1 ] && echo "Deployed." || echo "Dry run finished - nothing was changed. Re-run with --apply to deploy.")

  Site      : $SITE
  Logs      : aws logs tail $LOG_GROUP --follow
  API tasks : $CLUSTER / $SERVICE
$([ "$WITH_DEMO" -eq 1 ] && echo "  Demo user : demo@webond.test / Demo@1234  (publicly documented - dev only)")
Next: confirm the alarm e-mail subscription, and see docs/DEPLOYMENT.md (after-deploy checklist).
DONE
