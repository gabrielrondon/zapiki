# Advisor Meeting Gate Tasks

Objetivo: chegar na proxima reuniao com sinais fortes de aprovacao tecnica e de negocio.

## Gate A - Seguranca e trust model
- [x] A1. Politica de input privado implementada em runtime (`STORE_INPUT_DATA`, `MAX_INPUT_BYTES`, `MAX_PUBLIC_INPUT_BYTES`).
- [x] A2. Testes de regressao para no-persist e limites de payload.
- [x] A3. Trust model documentado com limites reais da arquitetura atual.
- [x] A4. Definir estrategia de single-tenant/on-prem para clientes regulados.
  - Progresso: estrategia inicial documentada em `docs/SINGLE_TENANT_DEPLOYMENT_STRATEGY.md`.

## Gate B - Produto pagavel
- [x] B1. SKUs e monetizacao documentados (`docs/PRICING_SKUS.md`).
- [ ] B2. 3 design partners com dor validada e criterio de sucesso por piloto.
- [ ] B3. Dashboard cliente (uso/falhas/custos) pronto para pilotos.

## Gate C - Operacao em escala
- [x] C1. CI verde com contract sync backend/frontend.
- [x] C2. Idempotencia worker + compensacao em falhas de enqueue.
- [x] C3. SLO com alertas ativos no ambiente de producao.
  - Progresso: watchdog ativo em producao via GitHub Actions (`.github/workflows/slo-watchdog.yml`) com execucao recorrente e canario de firing/clear por `workflow_dispatch`.
- [x] C4. Evidencia de carga (teste com volume alvo por tier).
  - Progresso: `load-evidence` executado com sucesso (60 req, concorrencia 10) em `artifacts/load-evidence-20260226-164042.md`:
    - HTTP 2xx: 60/60
    - non-2xx: 0
    - proof completed: 60
    - p95: 199.023ms

## Critério de convocacao da reuniao
Convocar quando: todos os itens `A*` estiverem fechados e pelo menos 2 itens `B*` + 3 itens `C*` estiverem fechados.
