# Infraestructura real — estado al 2026-10-19

**Este documento describe lo que existe, no lo ideal.** La Fase 0 del plan de lanzamiento
pide no asumir topología; a la fecha de este corte, la topología de producción **no está
decidida**. Este documento deja constancia de las opciones y de un hallazgo nuevo que debe
pesar en la decisión.

## Decisión pendiente: ¿dónde corre producción?

No hay VPS ni portátil designado todavía como host de producción. Las tres opciones y sus
compromisos reales:

| Opción | Qué implica | Riesgo principal |
|---|---|---|
| VPS nuevo aloja todo (DB + microservicios) | Un solo proveedor cloud, IP pública estable, posibilidad de SLA | Costo recurrente; hay que dimensionar CPU/RAM para 12 servicios + Postgres + Redis + RabbitMQ + Weaviate + Neo4j + EMQX a la vez |
| VPS solo datos (Postgres/Weaviate/Redis) + portátil corre microservicios | Más barato al inicio | El portátil es un punto único de falla: si se apaga, se suspende o pierde la IP, el sistema completo cae aunque los datos sigan vivos en el VPS |
| Portátil aloja todo | Costo cero de infraestructura | Combina el riesgo anterior con que los datos también viven en una máquina sin redundancia, sin UPS garantizado, con IP residencial dinámica (rompe DNS/SSL si cambia) |

**Recomendación técnica, no de negocio:** para un sistema que procesará datos de salud
(Ley 1581) y pagos reales (Wompi), un portátil como servidor permanente no es defendible
ante una auditoría de protección de datos ni ante un cliente que pregunte por disponibilidad.
Es una decisión de negocio válida para una fase de piloto cerrado con pocos usuarios, pero
debe tomarse sabiendo esto explícitamente, no por omisión.

## Hallazgo nuevo relevante para la decisión

La máquina usada en esta sesión de auditoría (Kali, 6 CPU, ~23 GiB RAM) **ya aloja otro
proyecto no relacionado en Docker** ("lucy-v6" / "nova"), con ~25 contenedores corriendo
de forma persistente, ocupando los puertos 5432 (Postgres), 6379 (Redis) y el rango
8700-8800. Si esta misma máquina llegara a considerarse como candidata ("el portátil"),
hay que saber que no está libre: cualquier despliegue de ZU360 aquí chocaría de entrada
con esos puertos y tendría que convivir con la carga de CPU/RAM de ese otro sistema.
Esto no es un problema del código de ZU360; es un dato de la máquina física.

## Dominio, DNS, SSL

**No verificado en esta sesión** — no se proporcionó un nombre de dominio para resolver con
`dig`/`nslookup`. Pendiente: nombre de dominio real, confirmación de registros A/AAAA/CNAME
apuntando a la IP del host que se elija, y decisión sobre Let's Encrypt u otro emisor de
certificado.

## Specs de máquinas candidatas

**No confirmadas.** Pendiente que el usuario indique CPU/RAM/disco/SO/IP pública de cada
máquina candidata (VPS nuevo y/o portátil) antes de proponer un pipeline de despliegue
(Fase 4).

## Qué se corrió hoy, y dónde

Todo el trabajo de verificación de este corte (Fase 1) se ejecutó **localmente, en Docker,
en esta misma máquina de auditoría** — no en ningún VPS ni en el portátil candidato a
producción. Es la primera vez que el stack Core+Postgres+Redis+RabbitMQ corre y se prueba
contra datos reales (ver `ESTADO_SISTEMA.md`), pero esa evidencia es de **entorno local de
pruebas**, no de la infraestructura de producción, que sigue sin definir.
