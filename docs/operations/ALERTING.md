# Zapiki Alerting (SLO)

## Arquivo de regras
- `docs/operations/prometheus-alert-rules.yaml`

## Alertas obrigatorios
- `ZapikiHigh5xxRate`: disponibilidade (5xx > 2% por 10m).
- `ZapikiHighP95Latency`: latencia (p95 > 500ms por 10m).
- `ZapikiQueueErrorSpike`: confiabilidade async (erros de fila acima do baseline).
- `ZapikiProofSuccessRateLow`: sucesso de provas < 95% por 30m.

## Validacao local
```bash
make check-slo-alerts
```

## Watchdog ativo em producao
- Workflow: `.github/workflows/slo-watchdog.yml`
- Frequencia: a cada 15 minutos
- Alvo: `https://zapiki-production.up.railway.app`
- Canario de alerta: executar manualmente com `force_alert=true` para validar estado `firing` e depois `false` para `clear`.

## Integracao no Prometheus
1. Incluir o arquivo no `rule_files` do Prometheus.
2. Recarregar configuracao (`/-/reload`) ou reiniciar.
3. Confirmar no UI que as regras aparecem carregadas.

## Acao operacional minima
- `ZapikiHigh5xxRate` / `ZapikiHighP95Latency`: rollback ou reduzir carga, investigar deploy recente.
- `ZapikiQueueErrorSpike`: analisar worker logs e backlog da fila.
- `ZapikiProofSuccessRateLow`: identificar proof system/circuito afetado e desabilitar temporariamente se necessario.
