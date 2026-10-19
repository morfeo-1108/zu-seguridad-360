"""Tests del registry con handlers locales inertes; nunca llaman servicios externos."""
import asyncio

import pytest

from app.skills.registry import Skill, SkillRegistry, SkillResult


def run(coro):
    return asyncio.run(coro)


def test_runtime_registry_is_closed_by_default():
    reg = SkillRegistry()
    assert reg.get("create_incident") is None
    assert reg.tool_specs("profesional_sst") == []
    assert reg.tool_specs("trabajador") == []
    result = run(reg.execute("create_incident", {}, {"role": "profesional_sst"}))
    assert result.ok is False
    assert "no existe" in (result.error or "")


def test_registration_requires_explicit_experimental_opt_in():
    async def handler(args, ctx, http):
        return SkillResult(ok=True)

    reg = SkillRegistry()
    skill = Skill("local_noop", "No-op test", {}, handler)
    with pytest.raises(RuntimeError, match="disabled by default"):
        reg.register(skill)


def test_experimental_handler_obeys_role_and_confirmation():
    async def handler(args, ctx, http):
        return SkillResult(ok=True, data={"synthetic": True})

    reg = SkillRegistry(enable_experimental=True)
    reg.register(Skill(
        name="local_noop",
        description="Only a local test handler",
        parameters={"value": {"type": "string", "required": True}},
        handler=handler,
        requires_confirmation=True,
        required_role="profesional_sst",
    ))
    assert len(reg.tool_specs("profesional_sst")) == 1
    assert reg.tool_specs("trabajador") == []

    pending = run(reg.execute("local_noop", {"value": "synthetic"},
                              {"role": "profesional_sst"}, confirmed=False))
    assert pending.ok is False
    assert pending.requires_confirmation is True

    denied = run(reg.execute("local_noop", {"value": "synthetic"},
                              {"role": "trabajador"}, confirmed=True))
    assert denied.ok is False
    assert "Permisos" in (denied.error or "")

    allowed = run(reg.execute("local_noop", {"value": "synthetic"},
                               {"role": "profesional_sst"}, confirmed=True))
    assert allowed.ok is True
    assert allowed.data == {"synthetic": True}
    run(reg.close())
