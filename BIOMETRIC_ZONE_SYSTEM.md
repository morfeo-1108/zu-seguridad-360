# Sistema de Biometría + Zonas + Validación por Dron (v6)

## El flujo de extremo a extremo

Un trabajador NO opera en un solo entorno. Una obra tiene zona de alturas, de
espacio confinado, de soldadura, eléctrica y general — cada una con protocolo
distinto. El sistema valida el checklist correcto contra el trabajador correcto
en la zona correcta.

```
Dron entra a una zona
   │
   ├─ 1. RECONOCIMIENTO DE ZONA  (core: ZoneService)
   │      • por marcador visual ArUco/QR físico en la zona, o
   │      • por GPS (point-in-polygon contra geocercas)
   │      → obtiene el CHECKLIST de esa zona
   │
   ├─ 2. INFERENCIA DE VISIÓN  (ai: VisionInferenceEngine, ONNX)
   │      • detecta EPP, caídas, fuego, humo, posturas
   │
   ├─ 3. BIOMETRÍA FACIAL  (ai: BiometricEngine, ArcFace)
   │      • identifica a CADA persona del frame (embeddings 512-d)
   │      • consentimiento Ley 1581 obligatorio en enrolamiento
   │
   ├─ 4. VALIDACIÓN POR TRABAJADOR  (ai: ComplianceValidator)
   │      • para cada trabajador identificado, evalúa el checklist de la zona
   │      • resultado: % cumplimiento, ítems faltantes, severidad
   │
   └─ 5. ESCALADO  (notifications + sla)
          • "Carlos Rodríguez en zona de alturas: falta arnés (CRÍTICO)
            y línea de vida (CRÍTICO)" → WhatsApp al supervisor
```

## Protocolos por zona (checklists normativos sembrados automáticamente)

| Zona | Norma | Ítems críticos |
|------|-------|----------------|
| Alturas | Res. 1409/2012 | casco+barbuquejo, arnés, eslinga, línea de vida, permiso |
| Espacio confinado | Res. 4276/2022 | monitor gases, arnés rescate, trípode, vigía, permiso |
| Soldadura | Res. 2400/1979 | careta, guantes carnaza, delantal, extintor cercano |
| Eléctrico | RETIE | guantes dieléctricos, casco clase E, calzado, LOTO |
| General | Res. 2400/1979 | casco, chaleco, botas punta de acero |

## Endpoints
- `POST /v1/projects/:id/zones` — crea zona, siembra checklist por norma
- `GET  /v1/projects/:id/zones/resolve-marker?marker_id=` — reconoce zona por marcador
- `GET  /v1/projects/:id/zones/resolve-location?lat=&lng=` — reconoce zona por GPS
- `POST /v1/zone-inspection/enroll-biometric` — enrola rostro (con consentimiento)
- `POST /v1/zone-inspection/inspect` — flujo completo de validación del dron

## Privacidad (Ley 1581 / Decreto 1377)
- El dato biométrico es SENSIBLE: requiere consentimiento explícito previo.
- Se almacena el EMBEDDING (vector), nunca la foto. No es reversible a imagen.
- Liveness (anti-spoofing) en enrolamiento evita suplantación.
