---
inclusion: always
---

## hands-on-project
- **ไม่ต้องแบ่ง unit of work** — ใช้ aidlc โหมด **comprehensive** หรือ **`quick`** เท่านั้น และ **ข้าม phase `decomposition`/โหมด incremental** (ไม่ซอยงานเป็น units)

## Pre-Implementation Check (บังคับ ทำก่อนทุก task)
- ก่อนจะดำเนินการ implement ในแต่ละ task **ต้องเช็คข้อมูล requirements และข้อกำหนด (spec/steering) ที่เกี่ยวข้องทุกครั้ง** เพื่อยืนยันว่าสิ่งที่จะทำตรงตาม scope ที่กำหนด
- หากพบว่ามีส่วนที่ต้องดำเนินการเพิ่มเติมนอกเหนือจาก scope/requirements ที่กำหนดไว้ **ต้องแจ้งผู้ใช้ก่อนทุกครั้ง** และรอการยืนยัน ก่อนที่จะดำเนินการใดๆ — ห้ามเพิ่มงานหรือฟีเจอร์เองโดยไม่แจ้ง

## Task Execution Order (ลำดับการรัน Task — บังคับ)

จัดลำดับความสำคัญในการรัน task ตามนี้เสมอ ห้ามข้ามลำดับ:

1. **Setup / โครงสร้างพื้นฐาน** — `git init -b main` ที่ root ของโปรเจกต์ (พร้อมสร้าง `.gitignore`), วางโครงโปรเจกต์, dependency, config, โครง layout พื้นฐาน ทำก่อนให้ครบถ้วนเสมอ
2. **Core features (Frontend + Mock data)** — พัฒนา Frontend โดยใช้ข้อมูลจำลองเป็นไฟล์ `.json` เพื่อให้เห็นหน้าเว็บจริงและปรับแต่ง UI/flow ได้ก่อน ก่อนจะเดินหน้า step ถัดไป
3. **Core features (Frontend + Backend/Service)** — เชื่อมต่อ backend จริง หรือ convert ข้อมูลจาก `.json` เป็น database จริง เฉพาะกรณีที่โปรเจกต์นั้นต้องใช้ database/backend จริง
4. **Step อื่น ๆ ตามลำดับ** — ทำหลังจากข้อ 1–3 เสร็จเรียบร้อยแล้ว เรียงตาม dependency
5. **ส่วนเสริม + Test / Verify** — ไว้ทีหลัง หรือเป็น optional
6. **Deploy** — optional ทำเป็นลำดับสุดท้าย

### กติกาการรัน (บังคับ)
- **ทำทีละ task** เรียงตาม dependency ห้ามข้ามลำดับ
- ก่อนทำแต่ละ task ต้องเช็ค spec/steering/requirements ที่ task นั้นอ้างอิงก่อน (ตาม Pre-Implementation Check) — **ห้ามเดา ห้ามทำเกิน scope**
- หากมี **การเปลี่ยนแปลง เพิ่มเติม หรืออัปเดตใด ๆ ที่อยู่นอกเหนือ scope ของ task/spec ที่กำลังทำ** — ไม่จำกัดเฉพาะ Design เช่น requirements, design, โครงสร้าง/ข้อมูล (data), ฟีเจอร์/ฟังก์ชัน, dependency/config หรือข้อมูลอื่นใดที่ task นั้นไม่ได้ระบุไว้ (task ไม่ได้ fix เฉพาะเรื่องใดเรื่องหนึ่ง) — **ต้องแจ้งผู้ใช้ก่อนทุกครั้ง** และรอยืนยัน หากตกลงให้ทำ ให้ **อัปเดตกลับเข้าไฟล์ที่เกี่ยวข้อง (requirements/design/spec/steering ฯลฯ) ภายใต้หัวข้อ `Add-on`** ทุกครั้ง
- เสร็จแต่ละ task ให้ติ๊ก `[x]` ใน `tasks.md` และอัปเดตไฟล์ติดตามสถานะ (ถ้ามี)
- จบแต่ละกลุ่ม ต้อง verify ว่า build/dev รันได้จริง ไม่มี error ค้าง
- **หยุดรอผู้ใช้ยืนยันเมื่อจบแต่ละกลุ่ม** ก่อนไปกลุ่มถัดไป
- ตรวจสอบความถูกต้องทุกครั้งก่อนส่งมอบ

### หมายเหตุความเข้ากันกับ aidlc (aidlc mapping)

ลำดับและกติกาข้างต้นเข้ากันได้กับ workflow ของ aidlc โดยให้ยึด mapping ต่อไปนี้เมื่อรันผ่าน aidlc:

- **"กลุ่ม" ใน 6 ลำดับ ≈ "execution wave/phase" ของ aidlc** — ลำดับ Setup → FE+mock → FE+backend ฯลฯ ถือเป็น project constraint ที่ป้อนให้ `aidlc-tasks` ใช้จัด wave ไม่ใช่สูตรที่ขัดกับการจัดลำดับตาม dependency ของ aidlc
- **Test/Verify เลื่อนไปท้าย/optional ได้ตามบริบท scope** — override ค่า default ของ `aidlc-implement` ที่ให้ test ผ่านทุก task ก่อนปิดงาน (ยึด Working Software / Definition of Done ด้านล่างเป็นเกณฑ์ verify ขั้นต่ำแทน)
- **เนื้อหาภายใต้หัวข้อ `Add-on`** — ให้ sync กลับเข้า artifact จริงของ aidlc (`requirements.md` / `design.md` / spec ที่เกี่ยวข้อง) และบันทึกใน `audit.md` ด้วย ไม่ใช่ append ลอย ๆ ท้ายไฟล์เดียว

## Working Software / Definition of Done (บังคับ)

- เว็บที่พัฒนาต้อง **ทำงานได้จริง (runnable)** — รัน dev server แล้วเปิดใช้งานได้จริงบน browser ไม่ใช่แค่ code ที่ compile ผ่าน
- ต้องทำ **ครบทุก function ที่กำหนดให้พัฒนา** ตาม scope งานที่กำหนดไว้ใน `requirements.md` ห้ามส่งงานที่ทำไม่ครบตาม scope
- **ห้ามรายงานว่า "เสร็จแล้ว" ถ้าระบบใช้งานจริงไม่ได้** — ทุกครั้งที่จะบอกว่าเสร็จ ต้องผ่านการตรวจสอบจริงก่อน
- ก่อนบอกว่าเสร็จ ต้อง verify: build/compile ผ่าน, dev server รันขึ้น, ทุก flow หลัก/ฟังก์ชันตาม scope คลิกใช้งานได้จริง, ไม่มี error ค้างใน console/terminal
- ถ้ายังมีส่วนที่ยังไม่เสร็จหรือใช้ไม่ได้ ให้ **แจ้งตามจริงว่าอะไรเสร็จ/ไม่เสร็จ** พร้อมสาเหตุ ห้ามกล่าวเกินจริง
- ถ้า verify ไม่ได้เพราะติดข้อจำกัดสภาพแวดล้อม (เช่น dependency/เครื่อง) ให้ระบุชัดว่าเป็นเพราะอะไร ห้ามสรุปว่าใช้งานได้โดยไม่ได้ทดสอบ
