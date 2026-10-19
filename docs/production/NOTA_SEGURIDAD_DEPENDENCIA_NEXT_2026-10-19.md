# Nota de seguridad de dependencia — Next.js

**Corte:** 19 de octubre de 2026. La instalación limpia detectó que `next@15.0.3` estaba vulnerable según el aviso oficial de Next.js para CVE-2025-66478, que identifica afectación de Next.js 15.x con App Router y marca `15.5.7` como versión corregida de la rama 15. [1]

El `package.json` se actualizó a `next@15.5.28`, última publicación estable de la rama 15 encontrada en el registro de npm durante este trabajo. La compatibilidad declarada para `15.5.28` incluye React `^19.0.0`; se conservaron `react@19.0.0` y `react-dom@19.0.0`, se alinearon sus tipos TypeScript y se actualizó Lucide a `1.59.0` para cumplir su peer range de React 19. [2] [3]

La elección de una versión posterior al umbral `15.5.7` corrige la vulnerabilidad indicada por ese aviso. Se fijó `postcss@8.5.31` también como override de dependencias transitivas. `npm audit --omit=dev` no encuentra vulnerabilidades de runtime. La auditoría completa todavía reporta 9 vulnerabilidades transitivas únicamente en herramientas de build/lint: 7 altas en `braces` y 2 moderadas en `postcss-selector-parser`. [4] [5]

Al corte, npm resuelve `braces@3.0.3`, mientras que el aviso reporta afectadas las versiones `<=3.0.3` y no publica una primera versión corregida. El parser de selectores se corrige en `7.1.6`, pero la cadena actual de Tailwind 3 solicita la serie 6.x. Tailwind recomienda una migración major a v4, con cambios de PostCSS/configuración y requisitos de navegador modernos; no se aplicó sin una validación visual y de compatibilidad del piloto. [6] [7] Esta deuda de dependencias dev sigue abierta y no se presenta el proyecto como aprobado para producción.

El lockfile debe conservarse. Antes de cualquier entrega se deben repetir `npm audit`, `npm run typecheck` y `npm run build`.

## References

[1]: https://nextjs.org/blog/CVE-2025-66478 "Next.js Security Advisory CVE-2025-66478"
[2]: https://www.npmjs.com/package/next "Next.js package releases on npm"
[3]: https://nextjs.org/blog/next-15 "Next.js 15 release and React compatibility"
[4]: https://github.com/advisories/GHSA-vfj7-8cjw-p6xm "GitHub advisory for braces stack exhaustion"
[5]: https://github.com/advisories/GHSA-rj75-hqrm-r3gf "GitHub advisory for postcss-selector-parser CPU exhaustion"
[6]: https://tailwindcss.com/docs/upgrade-guide "Tailwind CSS v3 to v4 upgrade guide"
[7]: https://tailwindcss.com/docs/installation/using-postcss "Tailwind CSS PostCSS installation"
