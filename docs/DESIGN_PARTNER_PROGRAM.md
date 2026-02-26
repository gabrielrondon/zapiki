# Design Partner Program (B2)

## Objetivo
Manter 3 design partners ativos com KPI de valor financeiro/operacional claro para cada piloto.

## Fonte de verdade
- Arquivo: `docs/design-partners/partners.csv`
- Relatorio: `make design-partner-report`

## Criterio de "ativo"
- `status=active`
- `kpi_target` preenchido
- `next_step` definido
- `last_update` atualizado

## Cadencia semanal
1. Atualizar `partners.csv` com status e proximo passo.
2. Executar `make design-partner-report`.
3. Publicar o markdown de evidencias em `artifacts/`.

## Comando
```bash
make design-partner-report
```
