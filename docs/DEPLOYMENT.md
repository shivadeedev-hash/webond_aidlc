# WeBond: คู่มือเอาขึ้น AWS

คู่มือนี้อธิบายการเอา WeBond ขึ้น AWS ด้วยสคริปต์ที่เตรียมไว้ใน `deploy/aws/`

> **สถานะ:** template และสคริปต์ผ่านการตรวจแบบไม่แตะ AWS แล้ว (cfn-lint, shellcheck, actionlint,
> dry run, build image ทั้ง arm64 และ amd64, เทสต์กับ MySQL 8.4) แต่**ยังไม่เคยรันกับบัญชี AWS จริง**
> การ deploy ครั้งแรกให้ใช้ `dev` ก่อนและอ่านผลทุกขั้น
>
> **มีค่าใช้จ่ายทันทีที่ deploy** แม้ไม่มีคนใช้: NAT Gateway, RDS, Application Load Balancer,
> Fargate, KMS, Secrets Manager, WAF, CloudWatch ประเมินค่าใช้จ่ายด้วย
> [AWS Pricing Calculator](https://calculator.aws/) ก่อน และลบด้วย `destroy.sh` เมื่อไม่ใช้

---

## 1. สิ่งที่จะถูกสร้าง

```
ผู้ใช้ --HTTPS--> CloudFront (+ security headers)
                   |-- /*      --> S3 (ไฟล์หน้าเว็บ, ปิดสาธารณะ เข้าได้ผ่าน CloudFront เท่านั้น)
                   '-- /api/*  --> Load Balancer (+ WAF) --> API บน ECS Fargate --> RDS MySQL
                                   รับเฉพาะ CloudFront       subnet ส่วนตัว         subnet ไม่มีทางออกเน็ต
```

| ส่วน | รายละเอียด |
|------|-----------|
| เครือข่าย | VPC 2 AZ แยก subnet สาธารณะ (load balancer) / app (API) / data (DB ไม่มี route ออกอินเทอร์เน็ต), VPC Flow Logs |
| หน้าเว็บ | S3 (เข้ารหัส, block public access, versioning) + CloudFront ใช้โดเมนเดียวกับ API จึงไม่มีปัญหา CORS |
| API | ECS Fargate (ARM64), root filesystem อ่านอย่างเดียว, auto scaling ตาม CPU, rollback อัตโนมัติถ้า deploy ไม่ผ่าน health check |
| ฐานข้อมูล | RDS MySQL 8.4 เข้ารหัสด้วย KMS, บังคับ TLS, รหัสผ่านจัดการโดย RDS ใน Secrets Manager (หมุนอัตโนมัติ), backup 7 วัน (prod 35 วัน + Multi-AZ + deletion protection), เก็บ snapshot สุดท้ายเมื่อลบ stack |
| ความปลอดภัย | WAF (rate limit + AWS managed rules: common, known bad inputs, SQLi), load balancer รับเฉพาะ IP ของ CloudFront และต้องมี header ลับ, security group อนุญาตเฉพาะทางที่จำเป็น, image scan on push, ECR tag ห้ามเขียนทับ |
| รหัสลับ | JWT secret สองตัวสร้างแบบสุ่มใน Secrets Manager ไม่อยู่ในโค้ดหรือไฟล์ใดๆ, task ของแอปไม่มีสิทธิ์ AWS ใดๆ |
| การเฝ้าระวัง | CloudWatch logs (เข้ารหัส, เก็บ 90 วัน), access log ของ load balancer ใน S3, alarm ส่งอีเมล (5xx, target ไม่ healthy, CPU API/DB, พื้นที่ DB) |

**ลำดับของ `deploy.sh`** (เพื่อไม่ให้โค้ดใหม่รันกับตารางเก่า):
1. สร้าง/อัปเดตโครงสร้างพื้นฐาน (ครั้งแรกเปิด API 0 ตัว)
2. build image API (ARM64) และ push ขึ้น ECR ด้วย tag ตาม git commit
3. รัน migration ด้วย image ใหม่เป็นงานครั้งเดียว ถ้าล้มจะหยุดทั้งหมด ไม่ปล่อยเวอร์ชันใหม่
4. ปล่อย API เวอร์ชันใหม่ และรอจนเสถียร
5. build หน้าเว็บ อัปโหลดขึ้น S3 และล้าง cache ของ CloudFront

---

## 2. ก่อนเริ่ม

1. **บัญชี AWS** และสิทธิ์สร้างทรัพยากรใน template (VPC, ECS, RDS, IAM role, KMS, CloudFront, WAF ฯลฯ)
   ครั้งแรกปกติใช้สิทธิ์ระดับผู้ดูแลบัญชี
2. **เครื่องมือในเครื่อง:** AWS CLI v2, Docker Desktop (เปิดอยู่), Node.js 20+
   ```bash
   aws --version
   docker ps
   node -v
   ```
3. **ล็อกอิน AWS** (เลือกแบบที่องค์กรใช้ แนะนำ SSO เพราะเป็นรหัสชั่วคราว):
   ```bash
   aws configure sso          # หรือ aws configure
   aws sts get-caller-identity   # ต้องเห็นเลขบัญชีของคุณ
   ```
4. **Region ที่รองรับ:** `ap-southeast-1` (ค่าเริ่มต้น), `ap-southeast-2`, `ap-northeast-1`, `us-east-1`,
   `us-west-2`, `eu-west-1` (จำกัดเพราะ access log ของ load balancer ต้องรู้บัญชีระบบของแต่ละ region)

---

## 3. Deploy ครั้งแรก (dev)

### 3.1 ดูแผนก่อน (ไม่เปลี่ยนอะไร)

```bash
./deploy/aws/deploy.sh --env dev
```

ทุกคำสั่งจะขึ้น `[dry-run]` อ่านให้แน่ใจว่าเป็นบัญชีและ region ที่ต้องการ

### 3.2 Deploy จริง

```bash
./deploy/aws/deploy.sh --env dev --apply --with-demo-data
```

- สคริปต์ขอให้พิมพ์ชื่อ stack (`webond-dev`) เพื่อยืนยัน
- `--with-demo-data` ใส่ข้อมูลสาธิตและบัญชี `demo@webond.test` / `Demo@1234`
  **รหัสผ่านนี้เปิดเผยในเอกสาร** ใช้กับ dev เท่านั้น (สคริปต์และ API ไม่ยอมใส่ให้ prod)
- ครั้งแรกใช้เวลานาน CloudFront และ RDS สร้างนานเป็นพิเศษ

เมื่อเสร็จ สคริปต์พิมพ์ `Site : https://xxxx.cloudfront.net` เปิดลิงก์นั้นได้เลย

> **ข้อจำกัดสำคัญของ dev ที่ไม่มีโดเมน:** ผู้ใช้กับ CloudFront คุยกันด้วย HTTPS แต่ CloudFront กับ
> load balancer คุยกันด้วย HTTP (เพราะไม่มีใบรับรองของคุณเอง) แม้จะจำกัดให้รับเฉพาะ CloudFront แล้ว
> ก็ห้ามใช้กับข้อมูลจริง ใช้ทดสอบเท่านั้น `prod` บังคับให้มีโดเมนและใบรับรอง (template ไม่ยอม deploy ถ้าไม่มี)

### 3.3 หลัง deploy

1. **ยืนยันอีเมล alarm** (ถ้าตั้ง `AlertEmail`): AWS ส่งอีเมลให้กดยืนยัน ไม่กดจะไม่ได้รับแจ้งเตือน
2. เปิดเว็บ สมัครสมาชิก กรอกรหัส 6 หลักจากอีเมล (ต้องตั้งอีเมลตามข้อ 4.4 ถ้าไม่ตั้ง dev จะสมัครได้แต่ยืนยันไม่ได้)
   ล็อกอิน ลองซื้อกองทุน ดูพอร์ต ส่งออก Excel
3. เปิด DevTools (F12) > Console ดูว่ามีคำเตือน `Content-Security-Policy-Report-Only` หรือไม่
   ตอนนี้ CSP เป็นโหมดรายงานอย่างเดียว เมื่อไม่มีคำเตือนแล้ว ค่อยเปลี่ยนเป็นบังคับใน `template.yaml`
   (`Content-Security-Policy-Report-Only` เป็น `Content-Security-Policy`)
4. ดู log: `aws logs tail /webond/dev/api --follow`

---

## 4. Production (ต้องมีโดเมน)

### 4.1 เตรียมโดเมนและใบรับรอง

สมมติใช้ `app.example.com`

| ใบรับรอง (ACM) | region | ครอบคลุม | ใช้กับ |
|---------------|--------|----------|--------|
| `CloudFrontCertificateArn` | **us-east-1 เท่านั้น** | `app.example.com` | ผู้ใช้ไปยัง CloudFront |
| `AlbCertificateArn` | region เดียวกับ stack | `origin.app.example.com` | CloudFront ไปยัง load balancer (HTTPS) |

ขอใบรับรองแบบ DNS validation ใน ACM ทั้งสองใบ ถ้าโดเมนอยู่ใน Route 53 ใส่ `HostedZoneId` แล้ว stack
สร้าง DNS record ให้เอง ถ้าไม่ใช่ ให้สร้างเองจาก output `DistributionDomainName` และ `OriginDnsTarget`

### 4.2 ไฟล์ค่าตั้ง

```bash
cp deploy/aws/params.example.env deploy/aws/params.prod.env
# แก้ DomainName, HostedZoneId, AlbCertificateArn, CloudFrontCertificateArn, AlertEmail
```

`params.*.env` ถูก gitignore ไว้ (มีเลขบัญชีและ ARN)

### 4.3 Deploy

```bash
./deploy/aws/deploy.sh --env prod            # ดูแผน
./deploy/aws/deploy.sh --env prod --apply    # deploy จริง
```

prod ไม่มีข้อมูลสาธิตและไม่มีบัญชีใดๆ ให้ผู้ใช้สมัครเองผ่านหน้าเว็บ

prod **บังคับให้ตั้ง `SmtpHost`** (ดูข้อ 4.4) ไม่อย่างนั้นผู้ใช้ใหม่ยืนยันอีเมลไม่ได้เลย

> **ยังไม่พร้อมสำหรับการใช้งานจริงกับเงินจริง:** ระบบชำระเงิน ราคา และกองทุนเป็นของจำลองทั้งหมด

### 4.4 อีเมล (ยืนยันบัญชี รีเซ็ตรหัสผ่าน แจ้งเตือน) แบบฟรี

API ส่งอีเมลผ่าน SMTP ใช้ Gmail ได้ฟรี (ประมาณ 500 ฉบับต่อวัน ไม่ต้องมีโดเมน) ไม่ใช้ Amazon SES เพราะ SES
คิดเงินหลังพ้นช่วงทดลอง และต้องยืนยันโดเมนกับขอออกจาก sandbox ก่อน

1. ในบัญชี Google เปิด 2-Step Verification แล้วสร้าง App password ที่
   <https://myaccount.google.com/apppasswords>
2. เก็บรหัสใน SSM Parameter Store แบบ SecureString (standard tier ฟรี) วิธีนี้รหัสไม่ไปอยู่ใน shell history:
   ```bash
   read -rs SMTP_PW      # วางรหัส 16 ตัว (ไม่มีช่องว่าง) แล้วกด Enter จอจะไม่แสดง
   aws ssm put-parameter --name /webond/prod/smtp-password --type SecureString --value "$SMTP_PW"
   unset SMTP_PW
   ```
3. ใส่ใน `deploy/aws/params.prod.env`:
   ```bash
   SmtpHost=smtp.gmail.com
   SmtpPort=465
   SmtpUsername=your.name@gmail.com
   SmtpPasswordParameter=/webond/prod/smtp-password
   ```
4. deploy ตามข้อ 4.3 ใน log ของ API จะมี `e-mail delivery enabled`

สิ่งที่ template ทำให้: เปิด outbound เฉพาะพอร์ต SMTP ที่ตั้ง (465 หรือ 587) ผ่าน NAT, ให้สิทธิ์ execution role
อ่านเฉพาะ parameter นั้น (`ssm:GetParameters`) และส่ง `APP_PUBLIC_URL` เป็นที่อยู่เว็บจริง (ใช้ในท้ายอีเมลแจ้งเตือน) อีเมลยืนยันและรีเซ็ตรหัสผ่านใช้รหัส OTP 6 หลัก ไม่มีลิงก์
ใช้ key `aws/ssm` ค่าเริ่มต้น ถ้าเข้ารหัส parameter ด้วย KMS key ของคุณเอง ต้องเพิ่ม `kms:Decrypt` ให้ role เอง

เปลี่ยนรหัส: สร้าง App password ใหม่ แล้ว `aws ssm put-parameter --overwrite ...` ตามด้วย
`aws ecs update-service ... --force-new-deployment` แล้วลบ App password เก่าในบัญชี Google

ข้อควรรู้: Gmail เหมาะกับปริมาณน้อย (ทดสอบหรือผู้ใช้กลุ่มเล็ก) อีเมลจะออกในชื่อ Gmail ของคุณ และอาจเข้าสแปม
ถ้าส่งไม่ผ่าน API จะลองใหม่อีก 4 ครั้ง (ภายในประมาณ 8 นาที) แล้วบันทึกสาเหตุไว้ในคอลัมน์ `last_error` ของ
ตาราง `notification_outbox` ถ้ารันหลาย task พร้อมกัน แต่ละอีเมลจะถูกส่งโดย task เดียวเท่านั้น

---

## 5. Deploy จาก GitHub Actions (ไม่บังคับ)

`.github/workflows/ci.yml` รันทุก push/PR (lint, เทสต์รวม MySQL, build image, ตรวจ template)
`.github/workflows/deploy.yml` รันเมื่อกดเอง (Actions > Deploy > Run workflow) ค่าเริ่มต้นเป็น dry run

ตั้งค่าครั้งเดียว:

1. ใน IAM สร้าง OIDC identity provider: URL `https://token.actions.githubusercontent.com`,
   audience `sts.amazonaws.com`
2. สร้าง IAM role ให้ GitHub assume ได้ (ไม่มี access key ถาวร) trust policy ตัวอย่าง
   (แทน `ACCOUNT_ID`, `OWNER/REPO`):
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Principal": { "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com" },
       "Action": "sts:AssumeRoleWithWebIdentity",
       "Condition": {
         "StringEquals": { "token.actions.githubusercontent.com:aud": "sts.amazonaws.com" },
         "StringLike": { "token.actions.githubusercontent.com:sub": "repo:OWNER/REPO:environment:*" }
       }
     }]
   }
   ```
   แนบสิทธิ์เท่าที่ deploy ต้องใช้ แยก role ของ dev กับ prod
3. ใน GitHub > Settings > Environments สร้าง `dev` และ `prod` (ใส่ required reviewers ที่ `prod`)
   แต่ละ environment ตั้ง variable `AWS_DEPLOY_ROLE_ARN` และถ้ามี `DEPLOY_PARAMS` = เนื้อหาไฟล์ params

---

## 6. งานดูแลระบบ

| งาน | วิธี |
|-----|------|
| ปล่อยเวอร์ชันใหม่ | commit แล้วรัน `deploy.sh --env <env> --apply` อีกครั้ง |
| ย้อนเวอร์ชัน | `git checkout <commit-เก่า>` แล้ว deploy (ECS ย้อนเองอัตโนมัติถ้าเวอร์ชันใหม่ไม่ผ่าน health check) |
| ดู log | `aws logs tail /webond/<env>/api --follow` |
| หมุน JWT secret | Secrets Manager > secret `JwtAccessSecret...` / `JwtRefreshSecret...` > Retrieve > Edit ใส่ค่าสุ่มใหม่ ≥ 32 ตัว แล้ว `aws ecs update-service --cluster webond-<env> --service webond-<env>-api --force-new-deployment` (ทุกคนจะถูกออกจากระบบ) |
| รหัสผ่านฐานข้อมูล | RDS หมุนให้อัตโนมัติ (ค่าเริ่มต้นทุก 7 วัน) แต่ task ที่รันอยู่ยังถือรหัสเก่า **หลังหมุนแต่ละครั้งต้องรัน** `aws ecs update-service --cluster webond-<env> --service webond-<env>-api --force-new-deployment` (ดูข้อ 7) |
| เปิด/ปิดตัวสร้างราคา (dev) | เรียก `POST /api/admin/price-generator` ด้วยบัญชีแอดมิน (ดู `scripts/live-prices.sh`) |
| ลบทั้งหมด | `./deploy/aws/destroy.sh --env dev` (ดูแผน) แล้วเติม `--apply` |

`destroy.sh` เก็บ snapshot สุดท้ายของฐานข้อมูลไว้ (ยังคิดเงินจนกว่าจะลบเอง) prod เปิด deletion protection
ต้องปิดเองก่อน:

```bash
aws rds modify-db-instance --db-instance-identifier <ชื่อ instance> --no-deletion-protection --apply-immediately
```

---

## 7. ข้อจำกัดและสิ่งที่ยังไม่ได้ทำ

- ยังไม่เคยรันกับ AWS จริง
- **การหมุนรหัสผ่านฐานข้อมูลยังไม่อัตโนมัติครบวง:** RDS หมุนรหัสเองทุก 7 วัน แต่ ECS ใส่รหัสให้ task
  ตอนเริ่มเท่านั้น และ API เปิด connection ใหม่ทุก 5 นาที ดังนั้นภายในไม่กี่นาทีหลังการหมุน API
  จะต่อฐานข้อมูลไม่ได้ (alarm 5xx จะดัง) จนกว่าจะ force new deployment ทางแก้ถาวรที่ยังไม่ได้ทำ:
  กฎ EventBridge ที่สั่ง redeploy เมื่อการหมุนสำเร็จ หรือให้ API อ่าน secret เองตอนต่อใหม่
- รันแค่ region เดียว ไม่มี disaster recovery ข้าม region
- ไม่มี GuardDuty / Security Hub / AWS Config ใน template นี้ (แนะนำเปิดระดับบัญชี)
- อีเมลใช้ Gmail SMTP (ฟรี) เหมาะกับปริมาณน้อย ยังไม่ได้ตั้ง SPF/DKIM ของโดเมนตัวเอง จึงส่งในนาม Gmail
- คำสั่งซื้อขายยังไม่สร้างการแจ้งเตือนหรืออีเมล
- CSP ยังเป็นโหมดรายงานอย่างเดียว (ดูข้อ 3.3)
- ไม่มี bastion หรือการเข้าเครื่องแบบ interactive (ตั้งใจ งานดูแลทำผ่าน one-off task เช่น `webond migrate`)
