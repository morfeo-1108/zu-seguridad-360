"""
Modelo predictivo de riesgo de accidentalidad por empresa.

DECISIÓN DE DISEÑO (honesta): esto NO reemplaza el cálculo GTC 45, que es
LEGAL y debe ser determinista (ND×NE=NP, NP×NC=NR). El GTC 45 se queda en
código determinista en el core.

Este modelo es COMPLEMENTARIO: predice la PROBABILIDAD de un accidente en los
próximos 90 días usando señales históricas, para que la profesional priorice
intervenciones preventivas. Es una ayuda a la decisión, no una decisión legal.

Por qué gradient boosting y no deep learning:
  - Datos tabulares + volumen moderado (decenas/cientos de empresas) → XGBoost/
    LightGBM supera a redes profundas y es INTERPRETABLE (SHAP), lo que importa
    cuando hay que explicarle a un cliente por qué su empresa es "alto riesgo".
  - Deep learning aquí sería sobre-ingeniería sin datos suficientes.

Se activa cuando N>50 empresas (suficiente señal). Antes: motor de reglas.
"""
from dataclasses import dataclass
from typing import Dict, List, Optional
import math


@dataclass
class RiskFeatures:
    """Features de entrada del modelo (todas derivables del core)."""
    ciiu_risk_class: int          # 1-5 clase ARL
    employee_count: int
    incidents_last_12m: int
    days_since_last_incident: int
    training_coverage_pct: float
    medical_exam_coverage_pct: float
    open_corrective_actions: int
    avg_risk_matrix_nr: float     # NR promedio de la matriz GTC 45
    epp_compliance_pct: float     # de las rondas de dron
    compliance_0312_pct: float


@dataclass
class RiskPrediction:
    probability_90d: float        # 0-1: prob. de AT en próximos 90 días
    risk_band: str                # bajo | medio | alto | crítico
    top_factors: List[Dict]       # explicabilidad (SHAP-like)
    recommendation: str


class RiskPredictor:
    """
    En producción: LightGBM entrenado por sector CIIU, con explicabilidad SHAP.
    Aquí: modelo lineal interpretable como baseline funcional hasta tener datos.
    """

    MIN_COMPANIES_FOR_ML = 50

    # Pesos del baseline (en producción los aprende el modelo por sector)
    WEIGHTS = {
        "ciiu_risk_class": 0.18,
        "incidents_last_12m": 0.22,
        "recency": 0.15,
        "training_gap": 0.12,
        "medical_gap": 0.10,
        "open_actions": 0.08,
        "matrix_nr": 0.10,
        "epp_gap": 0.10,
        "compliance_gap": 0.10,
    }

    def __init__(self, company_count: int = 0):
        self.company_count = company_count
        self.ml_active = company_count >= self.MIN_COMPANIES_FOR_ML

    def predict(self, f: RiskFeatures) -> RiskPrediction:
        # Normalización de features a [0,1] donde 1 = más riesgo
        signals = {
            "ciiu_risk_class": f.ciiu_risk_class / 5.0,
            "incidents_last_12m": min(f.incidents_last_12m / 10.0, 1.0),
            "recency": max(0, 1 - f.days_since_last_incident / 365.0),
            "training_gap": 1 - f.training_coverage_pct / 100.0,
            "medical_gap": 1 - f.medical_exam_coverage_pct / 100.0,
            "open_actions": min(f.open_corrective_actions / 20.0, 1.0),
            "matrix_nr": min(f.avg_risk_matrix_nr / 1500.0, 1.0),
            "epp_gap": 1 - f.epp_compliance_pct / 100.0,
            "compliance_gap": 1 - f.compliance_0312_pct / 100.0,
        }

        # Score lineal → logística para obtener probabilidad
        raw = sum(self.WEIGHTS[k] * signals[k] for k in self.WEIGHTS)
        probability = 1 / (1 + math.exp(-6 * (raw - 0.5)))

        # Explicabilidad: factores que más contribuyen
        contributions = sorted(
            [{"factor": k, "contribution": round(self.WEIGHTS[k] * signals[k], 3)}
             for k in self.WEIGHTS],
            key=lambda x: x["contribution"], reverse=True,
        )[:3]

        band = self._band(probability)
        return RiskPrediction(
            probability_90d=round(probability, 3),
            risk_band=band,
            top_factors=contributions,
            recommendation=self._recommend(band, contributions),
        )

    def _band(self, p: float) -> str:
        if p >= 0.75: return "crítico"
        if p >= 0.50: return "alto"
        if p >= 0.25: return "medio"
        return "bajo"

    def _recommend(self, band: str, factors: List[Dict]) -> str:
        if band in ("crítico", "alto"):
            top = factors[0]["factor"]
            actions = {
                "incidents_last_12m": "Priorizar investigación de causas raíz y cierre de correctivos.",
                "training_gap": "Cerrar brecha de capacitaciones obligatorias de inmediato.",
                "medical_gap": "Actualizar exámenes médicos ocupacionales vencidos.",
                "epp_gap": "Reforzar uso de EPP: intensificar rondas de dron y sensibilización.",
                "matrix_nr": "Revisar controles de los peligros con NR más alto (jerarquía de controles).",
                "compliance_gap": "Cerrar estándares mínimos pendientes de la Res. 0312.",
            }
            return actions.get(top, "Intervención preventiva prioritaria recomendada.")
        return "Mantener el programa preventivo actual. Riesgo bajo controlado."


def predict_risk(features: RiskFeatures, company_count: int = 0) -> RiskPrediction:
    return RiskPredictor(company_count).predict(features)
