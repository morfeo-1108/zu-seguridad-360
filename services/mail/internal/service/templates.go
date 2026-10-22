package service

// Plantillas SST. Las variables se escapan al rellenar el HTML (ver fill).
// No hay plantilla de "FURAT enviado a la ARL": hoy el FURAT se genera como
// borrador y no se transmite a ninguna ARL, así que un correo que dijera lo
// contrario sería falso.
var emailTemplates = map[string]struct{ Subject, Body string }{
	"welcome": {
		Subject: "Bienvenido a ZU Seguridad Integral 360 — {{company_name}}",
		Body:    `<h2>Bienvenido, {{contact_name}}</h2><p>La empresa <strong>{{company_name}}</strong> quedó registrada en ZU Seguridad Integral 360.</p><p>Profesional SST a cargo: <strong>{{sst_name}}</strong>.</p>`,
	},
	"cert_expiring": {
		Subject: "Certificado próximo a vencer — {{employee_name}}",
		Body:    `<p>El certificado de <strong>{{cert_type}}</strong> de <strong>{{employee_name}}</strong> vence el <strong>{{expiry_date}}</strong>.</p><p>Gestione la renovación antes de esa fecha.</p>`,
	},
	"incident_reported": {
		Subject: "Nuevo incidente reportado — {{company_name}}",
		Body:    `<p>Se reportó un incidente en <strong>{{company_name}}</strong>.</p><p><strong>Tipo:</strong> {{incident_type}}<br><strong>Descripción:</strong> {{description}}<br><strong>Fecha:</strong> {{date}}</p>`,
	},
	"sla_breach": {
		Subject: "Plazo vencido: {{event_label}} — {{company_name}}",
		Body:    `<p>Se venció el plazo de atención de <strong>{{event_label}}</strong> en <strong>{{company_name}}</strong>.</p><p><strong>Plazo:</strong> {{target}}<br><strong>Abierto desde:</strong> {{opened_at}}<br><strong>Referencia:</strong> {{resource_id}}</p><p>Ingrese a la plataforma para atenderlo.</p>`,
	},
	"notification": {
		Subject: "{{title}}",
		Body:    `<p>{{message}}</p>`,
	},
	"morning_briefing": {
		Subject: "Briefing SST del {{date}} — {{companies}} empresas",
		Body:    `<h2>Briefing del {{date}}</h2><p>Resumen generado automáticamente a partir de los registros de su cartera. No modifica ningún registro; cada acción requiere su revisión.</p>{{items}}`,
	},
	"monthly_report": {
		Subject: "Informe SST mensual — {{company_name}} — {{month}}",
		Body:    `<h2>Informe mensual SST</h2><p>Indicadores de {{month}}.</p><p><strong>IF:</strong> {{if_value}} | <strong>IS:</strong> {{is_value}} | <strong>Cumplimiento:</strong> {{compliance}}%</p>`,
	},
}
