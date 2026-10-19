---
name: zu360-correo-corp
description: >
  Skill para vincular y gestionar la cuenta corporativa de correo electrónico en ZU 360.
  Activa cuando el usuario mencione: vincular correo, configurar correo corporativo,
  cuenta de correo empresa, SMTP, Gmail corporativo, Outlook corporativo, Hotmail empresa,
  correo con dominio propio, @empresa.com, enviar correo desde Claude, conectar correo,
  autenticación de correo, configurar email, cuenta de Google Workspace, Microsoft 365,
  correo institucional, o cualquier tarea de vinculación o envío real de correos
  desde la plataforma ZU 360. También activa para "quiero que Claude envíe los correos
  directamente", "configura mi cuenta de correo", "conecta el email".
---

# ZU360 — Vinculación de Cuenta Corporativa de Correo

## Propósito
Conectar la cuenta de correo profesional de la SISO al sistema ZU 360 para que
Claude Code pueda enviar correos reales directamente — alertas, informes, notificaciones,
actas — sin que la profesional tenga que copiar y pegar en su cliente de correo.

---

## Proveedores Soportados

| Proveedor | Tipo | Puerto SMTP | Seguridad |
|---|---|---|---|
| **Google Workspace** (`@empresa.com` en Google) | OAuth2 / SMTP | 587 | TLS |
| **Gmail personal** (`@gmail.com`) | OAuth2 / SMTP | 587 | TLS |
| **Microsoft 365** (`@empresa.com` en Outlook) | OAuth2 / SMTP | 587 | TLS |
| **Outlook / Hotmail personal** | OAuth2 / SMTP | 587 | TLS |
| **Servidor propio** (cPanel, Zoho, otro) | SMTP directo | 587 / 465 | TLS/SSL |
| **GoDaddy / Hostinger email** | SMTP directo | 465 | SSL |

---

## Opción A — Google Workspace o Gmail (recomendado)

### Paso 1: Activar contraseña de aplicación (más seguro que usar la contraseña real)

```
1. Ir a: myaccount.google.com
2. Seguridad → Verificación en 2 pasos (debe estar activa)
3. Seguridad → Contraseñas de aplicaciones
4. Seleccionar: Aplicación = "Correo" | Dispositivo = "PC con Windows"
5. Copiar la contraseña de 16 caracteres generada
   → Esta es la contraseña que usará ZU 360, NO tu contraseña normal
```

### Paso 2: Configurar en ZU 360 (.env o config.json)

```env
# Archivo: C:\ZU360\.env  (nunca compartir este archivo)

MAIL_PROVIDER=gmail
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_SECURE=false
MAIL_USER=tu.correo@empresa.com
MAIL_PASS=abcd efgh ijkl mnop   ← contraseña de aplicación de 16 chars
MAIL_FROM_NAME=Nombre Apellido — SST
MAIL_FROM_EMAIL=tu.correo@empresa.com
MAIL_REPLY_TO=tu.correo@empresa.com
```

### Paso 3: Verificar conexión desde Claude Code

```powershell
claude "verifica la conexión SMTP con las credenciales en .env y envía
un correo de prueba a [tu-correo] con asunto 'Test ZU360 OK'"
```

---

## Opción B — Microsoft 365 / Outlook Corporativo

### Paso 1: Obtener credenciales SMTP de Microsoft 365

```
OPCIÓN B1 — SMTP básico (más simple):
  1. Ir a: admin.microsoft.com
  2. Configuración → Configuración de la organización → Correo
  3. Habilitar: "SMTP AUTH" para el usuario
  4. Usar:
     Host: smtp.office365.com
     Puerto: 587
     Usuario: tu.correo@empresa.com
     Contraseña: tu contraseña de Microsoft 365

OPCIÓN B2 — App Password (si tiene MFA activo):
  1. Ir a: mysignins.microsoft.com
  2. Información de seguridad → Agregar método → Contraseña de aplicación
  3. Copiar la contraseña generada
```

### Paso 2: Configurar en ZU 360

```env
MAIL_PROVIDER=outlook
MAIL_HOST=smtp.office365.com
MAIL_PORT=587
MAIL_SECURE=false
MAIL_USER=tu.correo@empresa.com
MAIL_PASS=tu-contraseña-o-app-password
MAIL_FROM_NAME=Nombre Apellido — SST
MAIL_FROM_EMAIL=tu.correo@empresa.com
```

---

## Opción C — Servidor Propio / Hosting (cPanel, Zoho, GoDaddy)

```env
MAIL_PROVIDER=smtp
MAIL_HOST=mail.tudominio.com        ← revisar en cPanel → Cuentas de Email
MAIL_PORT=465
MAIL_SECURE=true                    ← true para SSL (puerto 465)
MAIL_USER=sst@tudominio.com
MAIL_PASS=tu-contraseña-del-correo
MAIL_FROM_NAME=SST — Nombre Empresa
MAIL_FROM_EMAIL=sst@tudominio.com
```

---

## Archivo .env — Seguridad Obligatoria

```
REGLAS DE SEGURIDAD DEL ARCHIVO .env:

✅ Guardar SOLO en: C:\ZU360\.env
✅ Nunca subir a Google Drive, OneDrive ni compartir por WhatsApp
✅ Agregar al .gitignore si se usa Git
✅ Nunca pegar en chats, correos ni documentos
✅ Si se compromete: ir inmediatamente a Google/Microsoft y revocar
   la contraseña de aplicación y generar una nueva

PERMISOS DEL ARCHIVO (Windows PowerShell):
icacls "C:\ZU360\.env" /inheritance:r /grant:r "$env:USERNAME:(R,W)"
→ Solo tu usuario puede leer y escribir el archivo
```

---

## Módulo de Envío — Lo que Claude Code puede enviar automáticamente

Una vez configurado, Claude Code puede enviar directamente:

```
ENVÍOS AUTOMÁTICOS (Autopilot):
  ✉️  Alerta vencimiento certificado → trabajador + supervisor
  ✉️  Alerta vencimiento examen médico → empleador
  ✉️  Reporte mensual de indicadores SST → gerencia (día 5 de cada mes)
  ✉️  Acta COPASST generada → miembros del comité
  ✉️  Notificación de AT → ARL + gerencia (dentro de 2 horas)
  ✉️  Plan de acción pendiente → responsables con fecha límite
  ✉️  Convocatoria a capacitación → listado de asistentes

ENVÍOS BAJO SOLICITUD:
  ✉️  Cualquier documento generado → destinatario específico
  ✉️  Informe de inspección con fotos → cliente
  ✉️  Propuesta comercial → empresa prospecto
  ✉️  Respuesta a correo recibido → con contexto SST

FLUJO DE CONFIRMACIÓN (siempre):
  Claude redacta el correo
        ↓
  Muestra a la profesional: asunto + destinatario + cuerpo
        ↓
  Pregunta: "¿Envío este correo? (sí/no/editar)"
        ↓
  Solo envía con confirmación explícita
```

---

## Firma Corporativa Automática

Configurar una vez, se aplica a todos los correos:

```
# En C:\ZU360\config\firma.txt

---
[NOMBRE COMPLETO]
Profesional en Salud Ocupacional
Licencia SST N°: [XXX] | Vigencia: [DD/MM/AAAA]

📞 [Teléfono] | ✉️ [correo@empresa.com]
🌐 [sitio web si tiene]

[Nombre empresa o "Servicios independientes SST"]
Reg. 0312/2019 · GTC-45 · Dec. 1072/2015

CONFIDENCIALIDAD: Este correo contiene información de salud ocupacional
de carácter confidencial conforme a la Ley 1581/2012.
```

---

## Verificación Final — Checklist de Vinculación

```
□ Proveedor de correo identificado (Google / Microsoft / Propio)
□ Contraseña de aplicación generada (NO usar contraseña principal)
□ Archivo .env creado en C:\ZU360\ con permisos restringidos
□ Correo de prueba enviado y recibido exitosamente
□ Firma corporativa configurada en config\firma.txt
□ Confirmado que el correo llega como [nombre] no como "noreply"
□ Carpeta SMTP habilitada en el proveedor (si era necesario)
□ .env agregado al .gitignore (si usa control de versiones)
```

---

## Diagnóstico de Errores Comunes

| Error | Causa probable | Solución |
|---|---|---|
| `535 Authentication failed` | Contraseña incorrecta o no es app password | Regenerar contraseña de aplicación |
| `534 Application-specific password required` | MFA activo sin app password | Crear contraseña de aplicación |
| `Connection timeout` | Puerto bloqueado o host incorrecto | Verificar host y puerto con el proveedor |
| `550 Relay not permitted` | SMTP AUTH no habilitado | Habilitar en admin.microsoft.com |
| `Correo llega como spam` | Sin SPF/DKIM configurado | Configurar en el panel DNS del dominio |
| `Less secure app blocked` | Google bloqueó acceso básico | Usar OAuth2 o contraseña de aplicación |
