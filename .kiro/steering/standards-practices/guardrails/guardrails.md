---
inclusion: always
---

#Guardrails

- Source code references as the source of truth. 
- ห้ามมโน, ห้ามเดา, ห้ามเติมข้อมูลที่ไม่มีอยู่จริงใน source code โดยเด็ดขาด.
- ถ้าไม่มั่นใจ ห้ามเดา ลด Hallucination ต้องการให้ตอบแบบมีความรับผิดชอบ มีหลักฐานอ้างอิงที่เชื่อถือ ได้จาก Source code เท่านั้น.
- ทำการวิเคราะห์อย่างมีเหตุผล และเชื่อถือได้.

- ห้าม call soap api เพื่อ update menu ผ่าน api endpoint จริง (รวมถึง mock endpoint ภายนอกเช่น webhook.site) ให้ทดสอบด้วยการ print debug request และ payload ลง log แทน (SOAP_DRY_RUN) ไม่มี network call ออกนอก ไม่แตะเครื่อง SDK dev/prod จริง ข้ามได้เมื่อ user approve


# Refine spec rules
Specs are designed for continuous refinement, allowing you to update and enhance them as your project evolves. This iterative approach ensures that specifications remain synchronized with changing requirements and technical designs, providing a reliable foundation for development.
- Change request ให้ refine spec ก่อน แล้วค่อย implement
- Request update ให้ refine spec ก่อน แล้วค่อย implement
- Any implement up-to-date refine-spec-first