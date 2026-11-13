---
inclusion: always
---

# ASSESSMENT SCOPE
  - Deep Dive analysis source code แบบละเอียด line by line เพื่อสกัด knowledge ของ applications ให้ได้ information ของ application ที่ครบถ้วนสำหรับใช้ในการ implement และ improvement ต่อได้
  - หาไฟล์, components และ functions/methods ที่เกี่ยวข้องกัน แบบละเอียด
  - เปลี่ยน Source Code, เอกสาร, assets และฐานข้อมูล ให้กลายเป็น Knowledge Graph เพื่อให้ AI เข้าใจความสัมพันธ์ของทั้งโปรเจกต์ แทนการค้นหาแบบทีละไฟล์
  - Trace ความสัมพันธ์ระหว่าง API, Database และ Frontend แยกตาม Screen flow
  - Decompose ระบบ → Modularize Components → Map Relationships & Dependencies → Analyze Interconnections → Trace Screen Flow → Connect ระบบแบบ End-to-End.
  - เปลี่ยนโปรเจกต์ทั้งก้อนให้ AI เข้าใจได้ในไม่กี่นาที ช่วยลดเวลาการทำความเข้าใจระบบก่อนเริ่มพัฒนา
  - ช่วยวิเคราะห์และอธิบายโครงสร้างโปรเจกต์
 
# Guardrails
  - Source code references as the source of truth.    
  - ห้ามมโน, ห้ามเดา, ห้ามเติมข้อมูลที่ไม่มีอยู่จริงใน source code โดยเด็ดขาด.
  - ถ้าไม่มั่นใจ ห้ามเดา ลด Hallucination ต้องการให้ตอบแบบมีความรับผิดชอบ มีหลักฐานอ้างอิงที่เชื่อถือ ได้จาก Source code เท่านั้น.
  - ทำการวิเคราะห์อย่างมีเหตุผล และเชื่อถือได้.