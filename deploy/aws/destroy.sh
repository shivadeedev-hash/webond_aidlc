#!/usr/bin/env bash
#
# WeBond - tear down a deployed environment.
#
# SAFE BY DEFAULT: without --apply this only prints what it would do.
#
#   deploy/aws/destroy.sh --env dev                 # show the plan
#   deploy/aws/destroy.sh --env dev --apply         # delete (asks you to type the stack name)
#
# Options: --env dev|prod  --region REGION  --profile NAME  --apply  --yes  -h
#
# What happens
#   - the web and log buckets are emptied (CloudFormation cannot delete a non-empty bucket)
#   - the stack is deleted (API, load balancer, CloudFront, VPC, secrets...)
#   - the DATABASE IS KEPT AS A FINAL SNAPSHOT (DeletionPolicy: Snapshot), which keeps
#     costing a little until you delete it yourself.
#   - prod has RDS deletion protection on: disable it yourself first, on purpose:
#       aws rds modify-db-instance --db-instance-identifier <id> --no-deletion-protection
#
set -euo pipefail

ENVIRONMENT="dev"
REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-ap-southeast-1}}"
PROFILE=""
APPLY=0
ASSUME_YES=0

usage() { sed -n '2,19p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; }
die() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }
step() { printf '\n==> %s\n' "$*"; }

while [ $# -gt 0 ]; do
  case "$1" in
    --env) ENVIRONMENT="${2:-}"; shift 2 ;;
    --region) REGION="${2:-}"; shift 2 ;;
    --profile) PROFILE="${2:-}"; shift 2 ;;
    --apply) APPLY=1; shift ;;
    --yes) ASSUME_YES=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1 (try --help)" ;;
  esac
done
case "$ENVIRONMENT" in dev|prod) ;; *) die "--env must be dev or prod" ;; esac

STACK="webond-$ENVIRONMENT"
DRY=$((1 - APPLY))
AWS=(aws --region "$REGION")
[ -z "$PROFILE" ] || AWS+=(--profile "$PROFILE")

run() { if [ "$DRY" -eq 1 ]; then printf '    [dry-run] %s\n' "$*"; else "$@"; fi; }

output() {
  "${AWS[@]}" cloudformation describe-stacks --stack-name "$STACK" \
    --query "Stacks[0].Outputs[?OutputKey=='$1'].OutputValue" --output text 2>/dev/null || true
}

# Delete every object version and delete marker, in batches of 1000.
empty_versioned_bucket() {
  local bucket="$1" batch
  while :; do
    batch="$("${AWS[@]}" s3api list-object-versions --bucket "$bucket" --max-items 1000 \
      --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}' --output json 2>/dev/null || true)"
    case "$batch" in ''|*'"Objects": null'*|*'"Objects": []'*|'null') break ;; esac
    "${AWS[@]}" s3api delete-objects --bucket "$bucket" --delete "$batch" >/dev/null
  done
  while :; do
    batch="$("${AWS[@]}" s3api list-object-versions --bucket "$bucket" --max-items 1000 \
      --query '{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}' --output json 2>/dev/null || true)"
    case "$batch" in ''|*'"Objects": null'*|*'"Objects": []'*|'null') break ;; esac
    "${AWS[@]}" s3api delete-objects --bucket "$bucket" --delete "$batch" >/dev/null
  done
}

cat <<PLAN

WeBond AWS teardown
  environment : $ENVIRONMENT
  region      : $REGION
  stack       : $STACK
  mode        : $([ "$APPLY" -eq 1 ] && echo "APPLY (DELETES resources)" || echo "DRY RUN (changes nothing)")

PLAN

command -v aws >/dev/null 2>&1 || die "'aws' is not installed"

if [ "$DRY" -eq 1 ]; then
  step "Plan"
  echo "    1. empty the web bucket and the log bucket"
  echo "    2. aws cloudformation delete-stack --stack-name $STACK, then wait"
  echo "    3. the database is kept as a final snapshot"
  echo
  echo "Dry run finished - nothing was changed. Re-run with --apply to delete."
  exit 0
fi

"${AWS[@]}" sts get-caller-identity >/dev/null 2>&1 || die "AWS credentials are not configured"
"${AWS[@]}" cloudformation describe-stacks --stack-name "$STACK" >/dev/null 2>&1 \
  || die "stack $STACK does not exist in $REGION"

if [ "$ASSUME_YES" -eq 0 ]; then
  printf 'This DELETES the %s environment (the database is kept as a snapshot).\nType the stack name (%s) to continue: ' "$ENVIRONMENT" "$STACK"
  read -r answer
  [ "$answer" = "$STACK" ] || die "not confirmed; nothing was changed"
fi

step "Emptying buckets"
WEB_BUCKET="$(output WebBucketName)"
LOG_BUCKET="$(output LogBucketName)"
[ -z "$WEB_BUCKET" ] || { echo "web bucket: $WEB_BUCKET"; empty_versioned_bucket "$WEB_BUCKET"; }
[ -z "$LOG_BUCKET" ] || { echo "log bucket: $LOG_BUCKET"; run "${AWS[@]}" s3 rm "s3://$LOG_BUCKET" --recursive --only-show-errors; }

step "Deleting stack $STACK (this takes a while; CloudFront is slow to delete)"
run "${AWS[@]}" cloudformation delete-stack --stack-name "$STACK"
run "${AWS[@]}" cloudformation wait stack-delete-complete --stack-name "$STACK"

cat <<DONE

Deleted. Still billing a little until you remove it yourself:
  - the final database snapshot (aws rds describe-db-snapshots)
  - CloudWatch log groups outlive the stack only if you changed their retention policy
If deletion failed because the log bucket refilled (the load balancer was still writing), just re-run this script.
DONE
