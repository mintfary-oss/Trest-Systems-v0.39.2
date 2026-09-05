# Архитектура системы

## Развёртывание

```text
Пользователь
    |
    v
trestctl (Go)
    |
    v
Docker Compose
    |
    +-- postgres
    +-- redis
    +-- minio
    +-- nginx
    +-- api
    +-- web
    +-- worker
    +-- project-service
    +-- estimate-service
    +-- matching-service
    +-- supplier-service
    +-- quality-service
    +-- ai-service
```

## Существующие модули

- `modules/magasin-777` — marketplace/web/API.
- `modules/proektirovka-sdaniy` — архитектурные DXF-чертежи.
- `modules/super-sistema` — локальный AI-ассистент.
- `trest-installer` → `cmd/trestctl`.

## Данные

```text
.trest/
├── data/
│   ├── postgres/
│   ├── redis/
│   └── minio/
├── logs/
├── backups/
├── reports/
├── diagnostics/
├── state.json
└── config/
```

## Доменные события

`ProjectCreated`, `EstimateCalculated`, `EstimateApproved`, `OrderCreated`,
`ContractActivated`, `ContractorApplicationSubmitted`,
`ContractorApplicationApproved`, `SupplierApplicationSubmitted`,
`SupplierApplicationApproved`, `MaterialDeliveryScheduled`,
`ConstructionStarted`, `QualityReportSubmitted`, `OrderCompleted`,
`DisputeOpened`, `RatingSubmitted`.
