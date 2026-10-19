"""Registry for explicitly enabled experimental agent tools.

The production singleton is intentionally empty: no model-provided action can
trigger an API call until service-to-service auth, object authorization,
confirmation, and audit controls have been independently reviewed.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Awaitable, Callable

import httpx


@dataclass
class SkillResult:
    ok: bool
    data: dict[str, Any] = field(default_factory=dict)
    error: str | None = None
    requires_confirmation: bool = False


@dataclass
class Skill:
    name: str
    description: str
    parameters: dict[str, Any]
    handler: Callable[..., Awaitable[SkillResult]]
    requires_confirmation: bool = False
    required_role: str = "profesional_sst"

    def to_tool_spec(self) -> dict[str, Any]:
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": {
                    "type": "object",
                    "properties": self.parameters,
                    "required": [
                        key for key, value in self.parameters.items()
                        if value.get("required", False)
                    ],
                },
            },
        }


class SkillRegistry:
    """Registry starts closed; registration requires an explicit test/dev opt-in."""

    def __init__(self, *, enable_experimental: bool = False):
        self._skills: dict[str, Skill] = {}
        self._enable_experimental = enable_experimental
        self._http = httpx.AsyncClient(timeout=30.0)

    def register(self, skill: Skill) -> None:
        if not self._enable_experimental:
            raise RuntimeError("Executable skills are disabled by default")
        self._skills[skill.name] = skill

    def get(self, name: str) -> Skill | None:
        return self._skills.get(name)

    def tool_specs(self, role: str = "profesional_sst") -> list[dict[str, Any]]:
        return [
            skill.to_tool_spec()
            for skill in self._skills.values()
            if self._role_allows(role, skill)
        ]

    @staticmethod
    def _role_allows(role: str, skill: Skill) -> bool:
        hierarchy = {"administrador": 3, "profesional_sst": 2, "empleador": 1, "trabajador": 0}
        return hierarchy.get(role, 0) >= hierarchy.get(skill.required_role, 2)

    async def execute(
        self,
        name: str,
        args: dict[str, Any],
        ctx: dict[str, Any],
        confirmed: bool = False,
    ) -> SkillResult:
        skill = self._skills.get(name)
        if skill is None:
            return SkillResult(ok=False, error=f"Skill '{name}' no existe")
        if not self._role_allows(ctx.get("role", "trabajador"), skill):
            return SkillResult(ok=False, error="Permisos insuficientes para esta acción")
        if skill.requires_confirmation and not confirmed:
            return SkillResult(
                ok=False,
                requires_confirmation=True,
                data={"pending_action": name, "args": args},
            )
        try:
            return await skill.handler(args, ctx, self._http)
        except Exception:
            # Do not forward exception text, request headers, or upstream payloads.
            return SkillResult(ok=False, error="Error ejecutando la acción")

    async def close(self) -> None:
        await self._http.aclose()


# Safe default for the application runtime: no executable tools registered.
registry = SkillRegistry()
