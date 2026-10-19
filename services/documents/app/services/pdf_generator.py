"""
Motor de generación PDF para los 49 documentos SST normativos.
Usa ReportLab para docs estructurados y WeasyPrint para docs HTML.
"""
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import cm
from reportlab.lib.colors import HexColor, black, white
from reportlab.platypus import (
    SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle,
    HRFlowable, PageBreak, Image
)
from reportlab.lib.enums import TA_CENTER, TA_LEFT, TA_JUSTIFY
from reportlab.pdfgen import canvas
import qrcode
import io
from datetime import datetime
from typing import Dict, Any, Optional

# Colores corporativos ZU
ZU_BLUE    = HexColor("#2563EB")
ZU_DARK    = HexColor("#1C1C1A")
ZU_GRAY    = HexColor("#6B7280")
ZU_LIGHT   = HexColor("#F3F4F6")
ZU_SUCCESS = HexColor("#16A34A")
ZU_WARNING = HexColor("#D97706")
ZU_DANGER  = HexColor("#DC2626")


def generate_header_footer(canvas_obj, doc, company_name: str, doc_title: str, version: str = "v1.0"):
    """Encabezado y pie de página corporativo en cada hoja."""
    canvas_obj.saveState()

    # Header
    canvas_obj.setFillColor(ZU_BLUE)
    canvas_obj.rect(0, A4[1] - 2*cm, A4[0], 2*cm, fill=1, stroke=0)
    canvas_obj.setFillColor(white)
    canvas_obj.setFont("Helvetica-Bold", 14)
    canvas_obj.drawString(1*cm, A4[1] - 1.2*cm, doc_title)
    canvas_obj.setFont("Helvetica", 9)
    canvas_obj.drawRightString(A4[0] - 1*cm, A4[1] - 0.7*cm, company_name)
    canvas_obj.drawRightString(A4[0] - 1*cm, A4[1] - 1.2*cm, f"ZU Seguridad Integral 360 · {version}")

    # Footer
    canvas_obj.setFillColor(ZU_LIGHT)
    canvas_obj.rect(0, 0, A4[0], 1.5*cm, fill=1, stroke=0)
    canvas_obj.setFillColor(ZU_GRAY)
    canvas_obj.setFont("Helvetica", 7)
    canvas_obj.drawString(1*cm, 0.5*cm, f"Documento generado el {datetime.now().strftime('%d/%m/%Y %H:%M')} | Confidencial")
    canvas_obj.drawRightString(A4[0] - 1*cm, 0.5*cm, f"Página {doc.page}")

    canvas_obj.restoreState()


def generate_qr_code(data: str) -> io.BytesIO:
    """Genera código QR para verificación del documento."""
    qr = qrcode.QRCode(version=1, error_correction=qrcode.constants.ERROR_CORRECT_L, box_size=4, border=2)
    qr.add_data(data)
    qr.make(fit=True)
    img = qr.make_image(fill_color="black", back_color="white")
    buf = io.BytesIO()
    img.save(buf, format="PNG")
    buf.seek(0)
    return buf


class DocumentGenerator:
    """Generador principal para los 49 tipos de documentos SST."""

    GENERATORS = {
        "politica_sst": "_gen_politica_sst",
        "acta_copasst": "_gen_acta_copasst",
        "furat": "_gen_furat",
        "certificado_capacitacion": "_gen_certificado",
        "certificado_alturas": "_gen_certificado_alturas",
        "matriz_riesgos": "_gen_matriz_riesgos",
        "programa_capacitacion": "_gen_programa_capacitacion",
        "informe_indicadores": "_gen_informe_indicadores",
        "plan_emergencias": "_gen_plan_emergencias",
        "reglamento_higiene": "_gen_reglamento_higiene",
        "permiso_alturas": "_gen_permiso_alturas",
        "acta_entrega_epp": "_gen_acta_epp",
        "informe_investigacion": "_gen_informe_investigacion",
        "checklist_inspeccion": "_gen_checklist",
    }

    def generate(self, template_key: str, data: Dict[str, Any], doc_id: str) -> bytes:
        """Genera el PDF para el template especificado."""
        method_name = self.GENERATORS.get(template_key, "_gen_generic")
        method = getattr(self, method_name, self._gen_generic)
        return method(data, doc_id)

    def _gen_politica_sst(self, data: Dict, doc_id: str) -> bytes:
        buf = io.BytesIO()
        company = data.get("company_name", "Empresa")

        doc = SimpleDocTemplate(buf, pagesize=A4,
            rightMargin=2*cm, leftMargin=2*cm, topMargin=3*cm, bottomMargin=2.5*cm)

        styles = getSampleStyleSheet()
        title_style = ParagraphStyle("ZUTitle", parent=styles["Title"],
            fontSize=18, textColor=ZU_BLUE, spaceAfter=0.5*cm, alignment=TA_CENTER)
        heading_style = ParagraphStyle("ZUHeading", parent=styles["Heading1"],
            fontSize=13, textColor=ZU_DARK, spaceBefore=0.5*cm, spaceAfter=0.2*cm)
        body_style = ParagraphStyle("ZUBody", parent=styles["Normal"],
            fontSize=10, textColor=ZU_DARK, leading=16, alignment=TA_JUSTIFY)
        note_style = ParagraphStyle("ZUNote", parent=styles["Normal"],
            fontSize=8, textColor=ZU_GRAY, leading=12)

        story = []

        # Título
        story.append(Spacer(1, 0.5*cm))
        story.append(Paragraph("POLÍTICA DE SEGURIDAD Y SALUD EN EL TRABAJO", title_style))
        story.append(Paragraph(company.upper(), ParagraphStyle("co", parent=title_style, fontSize=14)))
        story.append(HRFlowable(width="100%", thickness=2, color=ZU_BLUE, spaceAfter=0.5*cm))

        # Metadatos
        meta_data = [
            ["NIT:", data.get("nit", ""), "Código CIIU:", data.get("ciiu", "")],
            ["Versión:", data.get("version", "1.0"), "Fecha:", datetime.now().strftime("%d/%m/%Y")],
            ["ARL:", data.get("arl", ""), "Vigencia:", "2 años"],
        ]
        meta_table = Table(meta_data, colWidths=[3*cm, 6*cm, 3*cm, 6*cm])
        meta_table.setStyle(TableStyle([
            ("FONTSIZE", (0,0), (-1,-1), 9),
            ("FONTNAME", (0,0), (0,-1), "Helvetica-Bold"),
            ("FONTNAME", (2,0), (2,-1), "Helvetica-Bold"),
            ("TEXTCOLOR", (0,0), (-1,-1), ZU_DARK),
            ("GRID", (0,0), (-1,-1), 0.25, ZU_LIGHT),
            ("BACKGROUND", (0,0), (0,-1), ZU_LIGHT),
            ("BACKGROUND", (2,0), (2,-1), ZU_LIGHT),
            ("PADDING", (0,0), (-1,-1), 5),
        ]))
        story.append(meta_table)
        story.append(Spacer(1, 0.5*cm))

        # Declaración principal
        story.append(Paragraph("DECLARACIÓN DE LA DIRECCIÓN", heading_style))
        policy_text = f"""
        <b>{company}</b>, en cumplimiento del Decreto 1072 de 2015 (Decreto Único Reglamentario del Sector Trabajo),
        la Resolución 0312 de 2019 y demás normas vigentes en materia de Seguridad y Salud en el Trabajo (SST),
        declara su compromiso con la protección de la salud y seguridad de todos sus trabajadores, contratistas
        y visitantes.
        <br/><br/>
        La empresa se compromete a implementar, mantener y mejorar continuamente el Sistema de Gestión de
        Seguridad y Salud en el Trabajo (SG-SST), asignando los recursos humanos, técnicos y económicos
        necesarios para prevenir los accidentes de trabajo y las enfermedades laborales.
        """
        story.append(Paragraph(policy_text, body_style))

        # Compromisos
        story.append(Paragraph("COMPROMISOS DE LA ORGANIZACIÓN", heading_style))
        commitments = [
            "Identificar los peligros, evaluar y valorar los riesgos presentes en los procesos de la empresa mediante la metodología GTC 45.",
            "Establecer los controles necesarios para prevenir los daños en la salud de los trabajadores.",
            "Cumplir la normatividad nacional vigente aplicable en materia de riesgos laborales.",
            "Investigar los accidentes de trabajo y enfermedades laborales ocurridos en la empresa.",
            "Proteger la seguridad y salud de todos los trabajadores, mediante la mejora continua del SG-SST.",
            f"Garantizar la participación de todos los trabajadores a través del COPASST en el mejoramiento del SG-SST.",
            "Asignar y comunicar las responsabilidades específicas en SST a todos los niveles de la organización.",
        ]
        for i, c in enumerate(commitments, 1):
            story.append(Paragraph(f"{i}. {c}", body_style))
            story.append(Spacer(1, 0.1*cm))

        # Objetivos SST
        story.append(Paragraph("OBJETIVOS DEL SG-SST", heading_style))
        objectives = [
            f"Alcanzar y mantener un cumplimiento ≥ 95% en los estándares mínimos de la Resolución 0312 de 2019.",
            "Reducir el Índice de Frecuencia (IF) de accidentes de trabajo en un 20% anual.",
            f"Garantizar la cobertura del 100% de los trabajadores en los programas de capacitación SST.",
            "Mantener la gestión de exámenes médicos ocupacionales conforme a la normativa vigente, validada por profesional competente antes de emitir este documento.",
        ]
        for obj in objectives:
            story.append(Paragraph(f"• {obj}", body_style))

        # Firmas
        story.append(Spacer(1, 1.5*cm))
        story.append(HRFlowable(width="100%", thickness=0.5, color=ZU_GRAY))
        story.append(Spacer(1, 0.3*cm))

        firma_data = [
            ["_" * 40, "", "_" * 40],
            [data.get("legal_rep", "Representante Legal"), "", data.get("sst_name", "Profesional SST")],
            ["Representante Legal", "", "Responsable SG-SST"],
            [f"C.C. {data.get('legal_rep_cc', '')}", "", f"Lic. SST {data.get('sst_license', '')}"],
        ]
        firma_table = Table(firma_data, colWidths=[8*cm, 2*cm, 8*cm])
        firma_table.setStyle(TableStyle([
            ("FONTSIZE", (0,0), (-1,-1), 9),
            ("FONTNAME", (0,1), (-1,1), "Helvetica-Bold"),
            ("ALIGN", (0,0), (-1,-1), "CENTER"),
            ("TEXTCOLOR", (0,0), (-1,-1), ZU_DARK),
        ]))
        story.append(firma_table)

        # QR verificación
        story.append(Spacer(1, 0.5*cm))
        qr_buf = generate_qr_code(f"https://app.zu360.co/verify/doc/{doc_id}")
        qr_img = Image(qr_buf, width=2*cm, height=2*cm)
        story.append(Table([[qr_img, Paragraph(f"<font size='7' color='#6B7280'>Verificar autenticidad:<br/>app.zu360.co/verify/doc/{doc_id[:8]}...</font>", note_style)]], colWidths=[2.5*cm, 15*cm]))

        doc.build(story, onFirstPage=lambda c, d: generate_header_footer(c, d, company, "POLÍTICA SST"),
                        onLaterPages=lambda c, d: generate_header_footer(c, d, company, "POLÍTICA SST"))
        return buf.getvalue()

    def _gen_acta_copasst(self, data: Dict, doc_id: str) -> bytes:
        """Genera acta de reunión COPASST."""
        buf = io.BytesIO()
        company = data.get("company_name", "Empresa")
        doc = SimpleDocTemplate(buf, pagesize=A4, rightMargin=2*cm, leftMargin=2*cm, topMargin=3*cm, bottomMargin=2.5*cm)
        styles = getSampleStyleSheet()
        title_s = ParagraphStyle("t", parent=styles["Title"], fontSize=16, textColor=ZU_BLUE, alignment=TA_CENTER)
        head_s  = ParagraphStyle("h", parent=styles["Heading2"], fontSize=12, textColor=ZU_DARK, spaceBefore=0.4*cm)
        body_s  = ParagraphStyle("b", parent=styles["Normal"], fontSize=10, leading=16, alignment=TA_JUSTIFY)

        story = [
            Spacer(1, 0.3*cm),
            Paragraph("ACTA DE REUNIÓN — COPASST", title_s),
            Paragraph(company, ParagraphStyle("co", parent=title_s, fontSize=12)),
            HRFlowable(width="100%", thickness=2, color=ZU_BLUE, spaceAfter=0.4*cm),
        ]

        # Datos de reunión
        reunion_data = [
            ["No. Acta:", data.get("acta_number", "001-2026"), "Fecha:", data.get("date", datetime.now().strftime("%d/%m/%Y"))],
            ["Hora inicio:", data.get("start_time", "09:00"), "Hora fin:", data.get("end_time", "10:00")],
            ["Lugar:", data.get("location", "Sala de reuniones"), "Ciudad:", data.get("city", "")],
        ]
        t = Table(reunion_data, colWidths=[3*cm, 6*cm, 3*cm, 6*cm])
        t.setStyle(TableStyle([
            ("FONTSIZE", (0,0), (-1,-1), 9), ("FONTNAME", (0,0), (0,-1), "Helvetica-Bold"),
            ("FONTNAME", (2,0), (2,-1), "Helvetica-Bold"), ("GRID", (0,0), (-1,-1), 0.25, ZU_LIGHT),
            ("BACKGROUND", (0,0), (0,-1), ZU_LIGHT), ("BACKGROUND", (2,0), (2,-1), ZU_LIGHT), ("PADDING", (0,0), (-1,-1), 5),
        ]))
        story += [t, Spacer(1, 0.3*cm)]

        # Asistentes
        story.append(Paragraph("1. ASISTENTES", head_s))
        asistentes = data.get("attendees", [{"name": "Por definir", "role": "Empleador", "signature": ""}])
        att_data = [["Nombre", "Rol", "Firma"]] + [[a.get("name",""), a.get("role",""), "_"*20] for a in asistentes]
        att_t = Table(att_data, colWidths=[7*cm, 5*cm, 6*cm])
        att_t.setStyle(TableStyle([
            ("BACKGROUND", (0,0), (-1,0), ZU_BLUE), ("TEXTCOLOR", (0,0), (-1,0), white),
            ("FONTNAME", (0,0), (-1,0), "Helvetica-Bold"), ("FONTSIZE", (0,0), (-1,-1), 9),
            ("GRID", (0,0), (-1,-1), 0.25, ZU_LIGHT), ("ROWBACKGROUNDS", (0,1), (-1,-1), [white, ZU_LIGHT]),
            ("PADDING", (0,0), (-1,-1), 5),
        ]))
        story += [att_t, Spacer(1, 0.3*cm)]

        # Orden del día y compromisos
        for section, content in [
            ("2. ORDEN DEL DÍA", data.get("agenda", ["Verificación compromisos anteriores", "Revisión indicadores SST", "Compromisos y cierre"])),
            ("3. COMPROMISOS ADQUIRIDOS", data.get("commitments", [{"action": "Por definir", "responsible": "", "date": ""}])),
        ]:
            story.append(Paragraph(section, head_s))
            if isinstance(content[0], str):
                for i, item in enumerate(content, 1):
                    story.append(Paragraph(f"{i}. {item}", body_s))
            else:
                comp_data = [["Compromiso", "Responsable", "Fecha límite"]] + [[c.get("action",""), c.get("responsible",""), c.get("date","")] for c in content]
                comp_t = Table(comp_data, colWidths=[9*cm, 4*cm, 4*cm])
                comp_t.setStyle(TableStyle([
                    ("BACKGROUND", (0,0), (-1,0), ZU_BLUE), ("TEXTCOLOR", (0,0), (-1,0), white),
                    ("FONTNAME", (0,0), (-1,0), "Helvetica-Bold"), ("FONTSIZE", (0,0), (-1,-1), 9),
                    ("GRID", (0,0), (-1,-1), 0.25, ZU_LIGHT), ("PADDING", (0,0), (-1,-1), 5),
                ]))
                story.append(comp_t)
            story.append(Spacer(1, 0.2*cm))

        doc.build(story, onFirstPage=lambda c, d: generate_header_footer(c, d, company, "ACTA COPASST"),
                        onLaterPages=lambda c, d: generate_header_footer(c, d, company, "ACTA COPASST"))
        return buf.getvalue()

    def _gen_certificado(self, data: Dict, doc_id: str) -> bytes:
        """Genera certificado de capacitación con QR verificable."""
        buf = io.BytesIO()
        c = canvas.Canvas(buf, pagesize=(29*cm, 21*cm))  # Landscape A4

        # Fondo
        c.setFillColor(HexColor("#FAFAF8"))
        c.rect(0, 0, 29*cm, 21*cm, fill=1, stroke=0)

        # Borde decorativo
        c.setStrokeColor(ZU_BLUE)
        c.setLineWidth(8)
        c.rect(0.5*cm, 0.5*cm, 28*cm, 20*cm, fill=0, stroke=1)
        c.setLineWidth(2)
        c.setStrokeColor(HexColor("#DBEAFE"))
        c.rect(0.8*cm, 0.8*cm, 27.4*cm, 19.4*cm, fill=0, stroke=1)

        # Header ZU
        c.setFillColor(ZU_BLUE)
        c.rect(0.5*cm, 17.5*cm, 28*cm, 3*cm, fill=1, stroke=0)
        c.setFillColor(white)
        c.setFont("Helvetica-Bold", 22)
        c.drawCentredString(14.5*cm, 19.2*cm, "ZU SEGURIDAD INTEGRAL 360")
        c.setFont("Helvetica", 11)
        c.drawCentredString(14.5*cm, 18.6*cm, "Plataforma Inteligente de Salud Ocupacional — Colombia")

        # Título certificado
        c.setFillColor(ZU_DARK)
        c.setFont("Helvetica-Bold", 28)
        c.drawCentredString(14.5*cm, 15.8*cm, "CERTIFICADO DE CAPACITACIÓN")

        # Decoración
        c.setStrokeColor(ZU_BLUE)
        c.setLineWidth(1.5)
        c.line(3*cm, 15.3*cm, 26*cm, 15.3*cm)

        # Otorga
        c.setFont("Helvetica", 13)
        c.setFillColor(ZU_GRAY)
        c.drawCentredString(14.5*cm, 14.5*cm, "Se certifica que")

        # Nombre empleado
        c.setFont("Helvetica-Bold", 22)
        c.setFillColor(ZU_BLUE)
        c.drawCentredString(14.5*cm, 13.3*cm, data.get("employee_name", "Nombre del Empleado").upper())

        # Identificación
        c.setFont("Helvetica", 11)
        c.setFillColor(ZU_GRAY)
        c.drawCentredString(14.5*cm, 12.6*cm, f"C.C. {data.get('document_number', '')}")

        # Curso
        c.setFont("Helvetica", 12)
        c.setFillColor(ZU_DARK)
        c.drawCentredString(14.5*cm, 11.6*cm, "Ha completado satisfactoriamente el curso de:")
        c.setFont("Helvetica-Bold", 16)
        c.setFillColor(ZU_DARK)
        c.drawCentredString(14.5*cm, 10.8*cm, data.get("training_title", "Curso de Capacitación SST"))

        # Datos curso
        details = [
            f"Empresa: {data.get('company_name', '')}",
            f"Duración: {data.get('duration_hours', '8')} horas",
            f"Fecha: {data.get('issued_at', datetime.now().strftime('%d/%m/%Y'))}",
            f"Norma: {data.get('norm_ref', 'Decreto 1072/2015')}",
        ]
        c.setFont("Helvetica", 10)
        c.setFillColor(ZU_GRAY)
        x_start = 4*cm
        for detail in details:
            c.drawString(x_start, 9.8*cm, detail)
            x_start += 6.5*cm

        # Vigencia
        c.setFillColor(ZU_SUCCESS)
        c.roundRect(10*cm, 8.5*cm, 9*cm, 0.8*cm, 4, fill=1, stroke=0)
        c.setFillColor(white)
        c.setFont("Helvetica-Bold", 11)
        c.drawCentredString(14.5*cm, 8.8*cm, f"Vigente hasta: {data.get('expires_at', '')}")

        # Folio
        c.setFont("Helvetica", 8)
        c.setFillColor(ZU_GRAY)
        c.drawCentredString(14.5*cm, 7.8*cm, f"Folio: {data.get('folio', doc_id)}")

        # Firmas
        for x, label, name in [(4*cm, "Representante Legal", data.get("legal_rep","")), (18*cm, "Profesional SST", data.get("sst_name",""))]:
            c.setStrokeColor(ZU_GRAY)
            c.setLineWidth(0.5)
            c.line(x, 6.5*cm, x + 7*cm, 6.5*cm)
            c.setFont("Helvetica-Bold", 9)
            c.setFillColor(ZU_DARK)
            c.drawCentredString(x + 3.5*cm, 6.1*cm, name)
            c.setFont("Helvetica", 8)
            c.setFillColor(ZU_GRAY)
            c.drawCentredString(x + 3.5*cm, 5.7*cm, label)

        # QR
        qr_buf = generate_qr_code(f"https://app.zu360.co/verify/cert/{data.get('folio', doc_id)}")
        c.drawImage(qr_buf, 12.5*cm, 3.5*cm, width=2.5*cm, height=2.5*cm, mask="auto")
        c.setFont("Helvetica", 7)
        c.setFillColor(ZU_GRAY)
        c.drawCentredString(14.5*cm, 3.2*cm, "Escanee para verificar autenticidad")

        c.save()
        return buf.getvalue()

    def _gen_furat(self, data: Dict, doc_id: str) -> bytes:
        buf = io.BytesIO()
        doc = SimpleDocTemplate(buf, pagesize=A4, rightMargin=1.5*cm, leftMargin=1.5*cm, topMargin=3*cm, bottomMargin=2*cm)
        styles = getSampleStyleSheet()
        title_s = ParagraphStyle("t", parent=styles["Title"], fontSize=14, textColor=ZU_BLUE, alignment=TA_CENTER)
        head_s  = ParagraphStyle("h", parent=styles["Normal"], fontSize=10, textColor=white, fontName="Helvetica-Bold")
        body_s  = ParagraphStyle("b", parent=styles["Normal"], fontSize=9, leading=14)

        story = [
            Paragraph("FORMULARIO ÚNICO DE REPORTE DE ACCIDENTE DE TRABAJO — FURAT", title_s),
            Paragraph("Ley 1562 de 2012 · Decreto 1530 de 1996", ParagraphStyle("sub", parent=styles["Normal"], fontSize=9, textColor=ZU_GRAY, alignment=TA_CENTER, spaceAfter=0.3*cm)),
            HRFlowable(width="100%", thickness=2, color=ZU_BLUE, spaceAfter=0.3*cm),
        ]

        sections = [
            ("I. DATOS DEL EMPLEADOR", [
                ["Razón Social:", data.get("company_name",""), "NIT:", data.get("nit","")],
                ["ARL:", data.get("arl",""), "Clase Riesgo:", data.get("risk_class","")],
                ["Dirección:", data.get("address",""), "Ciudad:", data.get("city","")],
            ]),
            ("II. DATOS DEL TRABAJADOR ACCIDENTADO", [
                ["Nombres:", data.get("employee_name",""), "Cédula:", data.get("document_number","")],
                ["Cargo:", data.get("position",""), "Tipo Contrato:", data.get("contract_type","")],
                ["EPS:", data.get("eps",""), "Salario Base:", data.get("salary","")],
            ]),
            ("III. DATOS DEL ACCIDENTE", [
                ["Fecha:", data.get("date",""), "Hora:", data.get("time","")],
                ["Lugar:", data.get("location",""), "Municipio:", data.get("municipality","")],
                ["Tipo AT:", data.get("at_type",""), "Parte Cuerpo:", data.get("body_part","")],
                ["Naturaleza Lesión:", data.get("injury_nature",""), "Mecanismo:", data.get("mechanism","")],
            ]),
        ]

        for section_title, rows in sections:
            # Header sección
            sec_t = Table([[Paragraph(section_title, head_s)]], colWidths=[17*cm])
            sec_t.setStyle(TableStyle([("BACKGROUND", (0,0), (-1,-1), ZU_BLUE), ("PADDING", (0,0), (-1,-1), 5)]))
            story.append(sec_t)

            row_t = Table(rows, colWidths=[3.5*cm, 5*cm, 3.5*cm, 5*cm])
            row_t.setStyle(TableStyle([
                ("FONTSIZE", (0,0), (-1,-1), 9), ("FONTNAME", (0,0), (0,-1), "Helvetica-Bold"),
                ("FONTNAME", (2,0), (2,-1), "Helvetica-Bold"), ("GRID", (0,0), (-1,-1), 0.25, ZU_LIGHT),
                ("ROWBACKGROUNDS", (0,0), (-1,-1), [white, ZU_LIGHT]), ("PADDING", (0,0), (-1,-1), 5),
            ]))
            story.append(row_t)
            story.append(Spacer(1, 0.2*cm))

        # Descripción detallada
        desc_t = Table([[Paragraph("IV. DESCRIPCIÓN DETALLADA DEL ACCIDENTE", head_s)]], colWidths=[17*cm])
        desc_t.setStyle(TableStyle([("BACKGROUND", (0,0), (-1,-1), ZU_BLUE), ("PADDING", (0,0), (-1,-1), 5)]))
        story.append(desc_t)
        desc_content = Table([[Paragraph(data.get("description", ""), body_s)]], colWidths=[17*cm])
        desc_content.setStyle(TableStyle([("BOX", (0,0), (-1,-1), 0.25, ZU_LIGHT), ("PADDING", (0,0), (-1,-1), 8), ("MINROWHEIGHT", (0,0), (-1,-1), 3*cm)]))
        story.append(desc_content)

        doc.build(story, onFirstPage=lambda c, d: generate_header_footer(c, d, data.get("company_name",""), "FURAT"),
                        onLaterPages=lambda c, d: generate_header_footer(c, d, data.get("company_name",""), "FURAT"))
        return buf.getvalue()

    def _gen_generic(self, data: Dict, doc_id: str) -> bytes:
        """Template genérico para documentos no implementados aún."""
        buf = io.BytesIO()
        doc = SimpleDocTemplate(buf, pagesize=A4)
        styles = getSampleStyleSheet()
        story = [
            Paragraph(data.get("title", "Documento SST"), styles["Title"]),
            Spacer(1, 0.5*cm),
            Paragraph(data.get("content", "Contenido del documento"), styles["Normal"]),
        ]
        doc.build(story)
        return buf.getvalue()

    # Aliases
    def _gen_certificado_alturas(self, data: Dict, doc_id: str) -> bytes:
        data["training_title"] = "Trabajo Seguro en Alturas — normativa vigente por validar"
        data["norm_ref"] = "Pendiente de validación normativa vigente"
        return self._gen_certificado(data, doc_id)

    def _gen_matriz_riesgos(self, data, doc_id): return self._gen_generic({"title": "Matriz de Riesgos GTC 45", **data}, doc_id)
    def _gen_programa_capacitacion(self, data, doc_id): return self._gen_generic({"title": "Programa Anual de Capacitaciones", **data}, doc_id)
    def _gen_informe_indicadores(self, data, doc_id): return self._gen_generic({"title": "Informe de Indicadores SST", **data}, doc_id)
    def _gen_plan_emergencias(self, data, doc_id): return self._gen_generic({"title": "Plan de Emergencias", **data}, doc_id)
    def _gen_reglamento_higiene(self, data, doc_id): return self._gen_generic({"title": "Reglamento de Higiene y Seguridad Industrial", **data}, doc_id)
    def _gen_permiso_alturas(self, data, doc_id): return self._gen_generic({"title": "Permiso de Trabajo en Alturas", **data}, doc_id)
    def _gen_acta_epp(self, data, doc_id): return self._gen_generic({"title": "Acta de Entrega de EPP", **data}, doc_id)
    def _gen_informe_investigacion(self, data, doc_id): return self._gen_generic({"title": "Informe de Investigación de Accidente", **data}, doc_id)
    def _gen_checklist(self, data, doc_id): return self._gen_generic({"title": "Checklist de Inspección de Seguridad", **data}, doc_id)
