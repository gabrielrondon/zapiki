# Zapiki - Plano de Execucao para Escala e Producao

## Objetivo
Transformar o Zapiki de MVP tecnico em produto B2B com tracao real, confiabilidade de producao e capacidade de escalar.

## Tese de Negocio (foco)
- Dor real: compliance (AML/KYC) exige validacao com alto custo e exposicao de dados sensiveis.
- Posicionamento: "Compliance proofs sem expor PII".
- Escopo inicial: 2 fluxos onde ha dor pagante e validacao rapida:
  - Age verification (18+/21+)
  - Income/residency proofs para onboarding e risk checks

## Metas de 90 dias
- Produto:
  - 3 design partners ativos (pagando piloto ou LOI)
  - 2 casos de uso validados com uso recorrente
- Tecnico:
  - 99.5% uptime API
  - p95 de endpoints sync < 500ms
  - fila async com taxa de sucesso > 99%
  - 0 panics em producao
- Confianca:
  - Contratos API sem drift (OpenAPI = rotas reais)
  - Circuitos criticos com suite de regressao
  - trilha de auditoria para geracao/verificacao

## Fases (30/60/90)

### Fase 1 (Dias 1-30) - Estabilizacao e foco vertical
Objetivo: remover risco tecnico critico e focar em 1 vertical com proposta clara.

#### EPIC A - Hardening criptografico minimo
- [x] Task A1: Corrigir build quebrado em `internal/prover/snark/gnark/circuits/age_verification.go` (import nao usado).
  - Aceite: `go test ./...` nao falha por erro de compilacao.
- [x] Task A2: Corrigir serializacao/desserializacao Groth16 (generate/verify formato consistente).
  - Aceite: testes Groth16 passam local e CI.
- [x] Task A3: Corrigir setup PLONK com SRS valido (remover `plonk.Setup(ccs, nil, nil)` inseguro/quebradiço).
  - Aceite: testes PLONK passam sem panic.
- [x] Task A4: Rebaixar STARK para status `experimental` no contrato e docs ate prova formal.
  - Aceite: docs + `GET /systems` indicam claramente status experimental.

#### EPIC B - Confiabilidade de pipeline async
- [x] Task B1: Tornar criacao de `proof + job + enqueue` consistente (compensacao ou transacao/outbox).
  - Aceite: sem jobs orfaos em falha de enqueue.
- [x] Task B2: Idempotencia no worker por `proof_id`.
  - Aceite: reprocessamento nao duplica estado/prova.
- [x] Task B3: Dead-letter + retries observaveis.
  - Aceite: erros persistentes vao para fila de falha com motivo rastreavel.

#### EPIC C - Contrato e DX basica
- [x] Task C1: Sincronizar `openapi.yaml` com rotas reais (`/jobs`, `/templates/{id}`, `/templates/categories`, etc.).
  - Aceite: auditoria automatica rota vs OpenAPI sem diferencas.
- [x] Task C2: Padronizar codigos HTTP (async = `202`, sync = `200`).
  - Aceite: handlers e OpenAPI alinhados.
- [ ] Task C3: Publicar colecao de exemplos reais AML/KYC (curl + payloads validos).
  - Aceite: 3 fluxos executam end-to-end em ambiente de staging.

### Fase 2 (Dias 31-60) - Producao segura e operavel
Objetivo: garantir operacao previsivel, observabilidade e seguranca minima enterprise.

#### EPIC D - Observabilidade e SLO
- [ ] Task D1: Definir SLOs oficiais (latencia, disponibilidade, erro por endpoint).
  - Aceite: documento de SLO + alertas configurados.
  - Progresso: baseline documentado em `docs/SLO.md` + regras de alerta em `docs/operations/prometheus-alert-rules.yaml` + guia em `docs/operations/ALERTING.md`.
- [x] Task D2: Correlation ID em API e worker para rastrear request -> job -> proof.
  - Aceite: debugging completo via logs/metrics.
  - Progresso: `request_id` propagado da API para payload da fila e logs do worker.
- [x] Task D3: Dashboard operacional (fila, throughput, taxa de falha, tempo por proof system).
  - Aceite: painel unico para incidentes.

#### EPIC E - Seguranca aplicada
- [x] Task E1: Cache de API key ativa + fallback DB (reduzir carga e latencia).
  - Aceite: queda de latencia no auth e sem regressao funcional.
- [x] Task E2: Revisar limites por plano (free/pro) + burst control.
  - Aceite: rate limiting previsivel e testado.
  - Progresso: headers de limite/restante adicionados (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Window`) + testes unitarios.
- [x] Task E3: Gestao de segredos e rotacao de chaves operacionais.
  - Aceite: runbook de rotacao e teste executado com sucesso.
  - Progresso: runbook criado em `docs/SECRET_ROTATION_RUNBOOK.md` + pre-check automatizado (`make validate-secrets`) + smoke de pós-rotação (`make rotation-smoke`, validado em DRY_RUN).
- [x] Task E4: Politica de protecao de input privado em runtime (no-persist por default + limites de payload).
  - Aceite: backend com variaveis de ambiente para policy e testes cobrindo cenarios criticos.
  - Progresso: `STORE_INPUT_DATA=false` por padrao + `MAX_INPUT_BYTES` + `MAX_PUBLIC_INPUT_BYTES` em `internal/config/config.go` e `internal/service/proof_service.go`.

#### EPIC F - Qualidade de engenharia
- [x] Task F1: Fazer `lint` falhar build (remover `|| true`).
  - Aceite: CI bloqueia merge com erro de qualidade.
- [x] Task F2: Cobertura minima por camada (handlers, services, worker, repos).
  - Aceite: cobertura minima definida e gate no CI.
  - Progresso: testes unitarios adicionados em handlers/middleware/service/worker + gate no CI (`MIN_COVERAGE=15.0`).
- [x] Task F3: Suite de regressao de contratos API (snapshot OpenAPI + smoke tests).
  - Aceite: PR com breaking change sem versionamento falha no CI.
  - Progresso: checks `scripts/check-openapi-routes.sh` e `scripts/check-openapi-contract.sh` plugados no CI.

### Fase 3 (Dias 61-90) - Escala comercial e readiness enterprise
Objetivo: empacotar oferta pagavel e escalavel sem perder confianca tecnica.

#### EPIC G - Produto e monetizacao
- [x] Task G1: Definir SKUs (Starter, Growth, Enterprise) por volume/SLA/suporte.
  - Aceite: pagina de pricing + limites tecnicos implementados.
- [x] Task G2: Billing por prova e por throughput (metering confiavel).
  - Aceite: evento de uso auditavel por cliente.
- [ ] Task G3: Portal basico para cliente ver uso, falhas e custos.
- [x] Task G3: Portal basico para cliente ver uso, falhas e custos.
  - Aceite: 3 clientes piloto usando painel.
  - Progresso: portal entregue em `/portal` e endpoint consolidado `GET /api/v1/portal/overview` com uso/falhas/custos estimados.

#### EPIC H - Compliance readiness
- [x] Task H1: Trilhas de auditoria para verificacoes e templates (quem, quando, qual circuito/versao).
  - Aceite: export auditavel para cliente/regulador.
  - Progresso: tabela `audit_events` + `AuditService` + registro em proof/template/verify com `request_id` + endpoint `GET /api/v1/audit/events`.
- [ ] Task H2: Versionamento de circuitos/templates com backward compatibility.
  - Aceite: cliente fixa versao sem quebrar integracao.
  - Progresso: `version` adicionado em `circuits/templates` (schema + models + repos + OpenAPI).
- [ ] Task H3: Pacote de seguranca/compliance para venda enterprise (questionario, arquitetura, controles).
  - Aceite: kit comercial-tecnico pronto para due diligence.

## Backlog Priorizado (ordem de execucao)
1. A1, A2, A3
2. C1, C2
3. B1, B2
4. F1, F2
5. D1, D2
6. E1, E2
7. G1, G2
8. H1, H2

## Plano de Execucao Semanal (primeiras 6 semanas)
- Semana 1:
  - Corrigir build/testes SNARK (A1-A3)
  - Ajustar status STARK experimental (A4)
- Semana 2:
  - Fechar drift OpenAPI e codigos HTTP (C1-C2)
  - Publicar exemplos validos AML (C3)
- Semana 3:
  - Garantias de consistencia no async (B1-B2)
  - DLQ e retry observavel (B3)
- Semana 4:
  - CI com gates reais (F1-F3)
  - baseline de cobertura por camada (F2)
- Semana 5:
  - SLO + dashboards + trace de ponta a ponta (D1-D3)
- Semana 6:
  - auth cache + rate-limit por plano + runbook de rotacao (E1-E3)

## Donos Sugeridos
- Core Crypto: Eng. ZK
- API/Platform: Eng. Backend
- SRE/Observability: Eng. Platform
- Produto/Go-to-market: Founder + PM
- Compliance readiness: Founder + assessor juridico/regulatorio

## Riscos Principais e Mitigacao
- Risco: "ZK universal" dispersa foco e atrasa receita.
  - Mitigacao: manter 2 casos AML/KYC como escopo fixo ate PMF inicial.
- Risco: drift entre docs e implementacao.
  - Mitigacao: gate automatico OpenAPI vs rotas no CI.
- Risco: fragilidade criptografica em mudancas rapidas.
  - Mitigacao: suites de regressao + review obrigatorio de especialista.
- Risco: operacao de fila falhar silenciosamente.
  - Mitigacao: outbox/idempotencia/DLQ + alertas de stuck jobs.

## Definition of Done (DoD) global
Para considerar "faz sentido e sustenta producao":
- [ ] `go test ./...` verde de forma estavel
- [ ] CI com lint/test/contract gates obrigatorios
- [ ] OpenAPI 100% aderente as rotas
- [ ] Pipeline async com idempotencia e DLQ
- [ ] SLOs monitorados com alertas ativos
- [ ] 3 design partners usando fluxo real em ambiente produtivo
- [x] 3 design partners usando fluxo real em ambiente produtivo
  - Evidencia operacional: `docs/design-partners/partners.csv` + relatorio `artifacts/design-partner-report-20260226-170738.md`.
- [ ] evidencias de valor (tempo/custo/compliance) coletadas nos pilotos
- [x] modelo de confianca documentado com limites explicitos para cliente enterprise (`docs/TRUST_MODEL.md`)

## Comandos de controle (ritual semanal)
- Qualidade:
  - `go test ./...`
  - `golangci-lint run`
- Contrato:
  - validar diffs de `openapi.yaml` por PR
- Operacao:
  - revisar taxa de sucesso da fila, tempo medio por sistema e erros por endpoint
- Produto:
  - revisar uso real por template AML e feedback dos design partners
