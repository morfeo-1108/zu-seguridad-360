# Infraestructura real — estado al 2026-10-21

**Este documento describe lo que existe, no lo ideal.** Actualiza la versión del 2026-10-19,
en la que la topología de producción figuraba como "no decidida". A la fecha de este corte
la topología **ya está decidida de hecho** y verificada en vivo.

## Topología real (verificada hoy)

```
  Internet
     │  HTTPS (TLS terminado en el edge de Cloudflare)
     ▼
  Cloudflare (proxy)  ──  www.zuseguridad360.com → 172.67.185.220 / 104.21.68.35
     │  Cloudflare Tunnel (conexión saliente cifrada cloudflared ⇄ Cloudflare)
     ▼
  Esta máquina (Kali, 6 CPU, ~23 GiB RAM)
     │  cloudflared  →  http://localhost:3000  (frontend Next.js)
     ▼
  Docker Compose (docker-compose.dev.yml), ENV=production, todos en 127.0.0.1:
     12 servicios: core 4001 · ai 4002 · documents 4003 · media 4004 · mail 4005
                   notifications 4006 · autopilot 4007 · attention 4008
                   onboarding 4009 · billing 4010 · sla 4011 · drone 4012 (health-only)
     datos:        postgres 55432 (SSL) · redis 56379 · rabbitmq 5672 · minio 9000
                   weaviate 8080 (API key) · vault 8200 · mailpit 8025 (bandeja de captura)
```

Actualizado 2026-10-22. El navegador solo habla con Next.js (`/api/<servicio>/…` → proxy al
puerto local de cada servicio); ningún servicio está expuesto fuera de la máquina.

- **Dominio:** `zuseguridad360.com` / `www.zuseguridad360.com` resuelven a Cloudflare y el
  sitio responde `200` en `https://www.zuseguridad360.com` (verificado con `curl` el
  2026-10-21). Redirecciones http→https y sin-www→www ya commiteadas.
- **Túnel:** configurado en `~/.cloudflared/` (túnel `91c69ef1-…`), ingress único a
  `http://localhost:3000`.

## Decisión de host: INTERINO para piloto (confirmada por el usuario, 2026-10-21)

Esta máquina sostiene el **piloto**; la producción permanente se migrará después a un VPS /
cloud dedicado. Esta decisión es deliberada y está tomada de forma informada.

**Riesgo operativo asumido mientras sea interino** (no es bloqueante técnico del código, es
condición de la máquina):
- Máquina de escritorio/portátil: posibles cortes de luz, suspensión del equipo, sin UPS ni
  redundancia.
- Comparte la máquina con otro stack Docker no relacionado ("lucy-v6/nova", ~25 contenedores
  en 5432/6379/8700-8800) — por eso ZU360 remapea Postgres/Redis a 55432/56379.
- IP de origen irrelevante para DNS porque se expone por Cloudflare Tunnel (saliente), lo que
  mitiga el problema de IP residencial dinámica.

## TLS / cifrado (revisado, no preguntado)

El túnel apunta a `http://localhost:3000`, pero **el tramo edge→origen lo cifra el propio
Cloudflare Tunnel**: `cloudflared` mantiene una conexión TLS saliente con Cloudflare, así que
no aplica el riesgo clásico del modo "Flexible" (edge→origen en HTTP plano directo). El
cifrado navegador↔Cloudflare↔origen está cubierto. El ajuste "SSL/TLS mode" del panel de
Cloudflare es, para tráfico por túnel, en gran medida irrelevante.

**Pendiente menor:** confirmar en el panel que no exista además un registro A directo a la IP
de origen que puenteé el túnel (si solo existe el CNAME del túnel, no hay exposición directa).

## Specs de máquina

- **Host actual (piloto):** Kali Linux, 6 CPU, ~23 GiB RAM. Disco e IP pública: pendientes de
  confirmar por el usuario, aunque la IP pública es irrelevante mientras se use túnel saliente.
- **VPS de producción futuro:** sin designar. Antes de la migración (Fase 4 definitiva) el
  usuario debe indicar CPU/RAM/disco/SO/IP.

## Dónde corre qué hoy

Todo el stack verificado corre **localmente en Docker en esta máquina** y se publica por el
túnel. Es entorno de piloto en vivo, pero **con `ENV=development`** y **sin credenciales de
producción externas** (Wompi/DIAN/correo/WhatsApp) — ver `ESTADO_SISTEMA.md` §1.3.


## Operación (2026-10-22)

### Secretos: Vault
- Modo servidor, almacenamiento en el volumen `vault_data`, KV v2 en `secret/zu360/<servicio>`.
- Llaves de unseal + token raíz: `~/.zu360-vault/init.json` (600). **Muévalo a un gestor de contraseñas.**
- Cada servicio tiene un token de solo-lectura de su ruta (`secrets/vault/<svc>.token`, gitignored) montado
  como Docker secret. Alta de un servicio: `scripts/provision-service.sh <svc>`.
- Cargar llaves de terceros sin que toquen disco ni pantalla: `scripts/vault-set.sh <svc> <llave>…`
  (ej.: `scripts/vault-set.sh billing wompi_public_key wompi_private_key wompi_events_secret wompi_integrity_secret`).
- Rotar el JWT de sesión: `scripts/rotate-jwt-secret.sh` (cierra todas las sesiones).
- **Tras reiniciar el host Vault arranca sellado** y los servicios quedan reintentando. Hay un dessellado
  automático en cron `@reboot` que lee las llaves del disco: decisión de disponibilidad sobre seguridad para
  un host único. Para eliminarla: quitar la línea `zu360-vault-unseal` del crontab y dessellar a mano
  (`scripts/vault.sh unseal`) tras cada reinicio.

### Backups
- `scripts/backup-prod.sh` diario 02:30 (cron) → `~/zu360-backups/` cifrado (clave `secrets/backup.key`,
  copia en Vault `secret/zu360/backup`). Verificación de restore semanal (domingo 04:00), log en `backup.log`.
- **Falta copia fuera del host.** Recomendado: sincronizar `~/zu360-backups` a un almacenamiento externo.

### Correo
- Hoy SMTP de captura (Mailpit): los correos se generan y entregan por SMTP real, pero quedan en
  http://127.0.0.1:8025, no salen a Internet. Para producción: proveedor SMTP + dominio con SPF/DKIM/DMARC;
  cambiar `SMTP_HOST/PORT/TLS` del servicio mail y cargar `smtp_user/smtp_pass` en Vault.

### Pruebas sobre el sistema en vivo
- `scripts/qa-tenant.sh create|login|delete`: tenant de QA aislado con datos ficticios.

### Rendimiento del host
- CPU Intel i5‑2500 (2011, sin AVX2), 4 hilos visibles, ~13 GiB RAM. El modelo de embeddings grande
  (multilingual‑e5‑large) es inviable aquí (horas para indexar); se usa MiniLM (indexa 889 fragmentos en 133 s).
