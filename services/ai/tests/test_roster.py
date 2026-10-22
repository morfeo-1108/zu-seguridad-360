"""Tests del roster de agentes — las dos clases de autonomía."""
import os
import sys

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.agents.roster import DETERMINISTIC_AGENTS, REASONING_AGENTS, AgentClass, roster_summary


def test_roster_has_both_classes():
    summary = roster_summary()
    assert summary["reasoning"] == len(REASONING_AGENTS)
    assert summary["deterministic"] == len(DETERMINISTIC_AGENTS)
    assert summary["total"] == summary["reasoning"] + summary["deterministic"]


def test_reasoning_agents_are_llm():
    for a in REASONING_AGENTS:
        assert a.cls == AgentClass.REASONING


def test_deterministic_agents_no_llm():
    """Los agentes deterministas NUNCA son LLM (facturar, vencimientos...)."""
    for a in DETERMINISTIC_AGENTS:
        assert a.cls == AgentClass.DETERMINISTIC


def test_billing_is_deterministic():
    """Crítico: facturación JAMÁS debe ser un agente de razonamiento (LLM)."""
    billing = [a for a in DETERMINISTIC_AGENTS if a.key == "billing"]
    assert len(billing) == 1
    assert billing[0].cls == AgentClass.DETERMINISTIC


def test_vision_agent_exists_and_reasoning():
    vision = [a for a in REASONING_AGENTS if a.key == "vision"]
    assert len(vision) == 1
    assert vision[0].cls == AgentClass.REASONING
