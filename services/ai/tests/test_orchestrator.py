"""Tests del ruteo del orquestador multi-agente."""
import sys, os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.agents.orchestrator import Orchestrator, AgentRole


def test_routes_incident_keywords():
    orch = Orchestrator()
    assert orch.route("Reportar un accidente de trabajo grave") == AgentRole.INCIDENT
    assert orch.route("Necesito generar el FURAT del incidente") == AgentRole.INCIDENT


def test_routes_compliance_keywords():
    orch = Orchestrator()
    assert orch.route("¿Qué certificados están por vencer?") == AgentRole.COMPLIANCE
    assert orch.route("Genera el acta del COPASST") == AgentRole.COMPLIANCE


def test_routes_vision_keywords():
    orch = Orchestrator()
    assert orch.route("Revisa la detección de EPP del dron") == AgentRole.VISION


def test_routes_normative_keywords():
    orch = Orchestrator()
    assert orch.route("¿Qué dice la resolución 1409 sobre alturas?") == AgentRole.NORMATIVE


def test_defaults_to_normative_when_ambiguous():
    """Sin palabras clave claras, cae al agente normativo (consulta general)."""
    orch = Orchestrator()
    assert orch.route("Hola, buenos días") == AgentRole.NORMATIVE
