# Доменные сущности

## User
`id`, `role`, `name`, `email`, `phone`, `status`, `createdAt`

## Organization
`id`, `type`, `legalName`, `registrationData`, `verificationStatus`, `rating`, `geography`

## Project
`id`, `customerId`, `objectType`, `location`, `parameters`, `architectureVersion`, `bimModelUrl`, `status`

## Estimate
`id`, `projectId`, `version`, `items`, `laborCost`, `materialsCost`, `logisticsCost`, `totalCost`, `currency`, `validUntil`, `status`

## Order
`id`, `projectId`, `customerId`, `approvedEstimateId`, `constructionStartDate`, `constructionEndDate`, `paymentMethod`, `status`, `contractId`

## ContractorApplication
`id`, `orderId`, `contractorId`, `proposedPrice`, `proposedDates`, `status`, `reviewReason`

## SupplierApplication
`id`, `orderId`, `supplierId`, `materials`, `deliveryDate`, `totalPrice`, `certificates`, `status`

## Rating
`id`, `authorId`, `subjectId`, `orderId`, `score`, `criteria`, `comment`

## QualityReport
`id`, `orderId`, `authorId`, `photos`, `videos`, `bimComparison`, `aiAnalysis`, `humanReview`, `status`

## Stage 4 — Constructor entities
- `ObjectType`: `id`, `code`, `name`, `description`, `parametersSchema`, `active`
- `Material`: `id`, `code`, `name`, `category`, `unit`, `parameters`, `active`
- `EngineeringSystem`: `id`, `code`, `name`, `category`, `parameters`, `active`
- `Finish`: `id`, `code`, `name`, `category`, `materialId`, `parameters`, `active`
- `ProjectVersion`: `id`, `projectId`, `version`, `status`, `snapshot`, `createdBy`, `createdAt`, `approvedAt`
- `ProjectMaterial`: `projectId`, `materialId`, `quantity`, `parameters`
- `ProjectEngineeringSystem`: `projectId`, `engineeringSystemId`, `parameters`
- `ProjectFinish`: `projectId`, `finishId`, `room`, `quantity`, `parameters`

## AIRecommendation
`id`, `agentType`, `entityType`, `entityId`, `inputSnapshot`, `recommendation`, `confidence`, `modelVersion`, `humanDecision`


## Constructor entities
`ObjectType`: `id`, `code`, `name`, `description`, `parametersSchema`, `active`

`Material`: `id`, `code`, `name`, `category`, `unit`, `parameters`, `active`

`EngineeringSystem`: `id`, `code`, `name`, `category`, `parameters`, `active`

`Finish`: `id`, `code`, `name`, `category`, `materialId`, `parameters`, `active`

`ProjectVersion`: `id`, `projectId`, `version`, `status`, `snapshot`, `createdBy`, `createdAt`, `approvedAt`

`ProjectMaterial`: `projectId`, `materialId`, `quantity`, `parameters`

`ProjectEngineeringSystem`: `projectId`, `engineeringSystemId`, `parameters`

`ProjectFinish`: `projectId`, `finishId`, `room`, `quantity`, `parameters`


## Stage 5 — Estimates and Orders
- `Estimate` belongs to a project/object and has a lifecycle independent from its immutable versions.
- `EstimateVersion` stores a versioned calculation snapshot and line items.
- `EstimateVersionItem` stores quantity, unit price and calculated total for a specific estimate version.
- `Order` references an approved estimate version where applicable, has planned dates, optional contractor/supplier assignment, and follows the controlled state machine.
- Financial/order creation supports an optional idempotency key with a unique database constraint.


## Stage 7 — Suppliers
- `supplier_profiles` — supplier organization profile, categories, delivery regions/terms, verification and active state.
- `supplier_applications` — supplier onboarding application and review status.
- `supplier_offers` — SKU/product offer, unit, price/currency, MOQ, stock, lead time and publication status.
- `supplier_documents` — certificate/document metadata and expiration.
- `supplier_offer_id` on estimate items/orders links commercial selection to marketplace data.

## BIM domain
`BIMModel` belongs to a project and may link to a project architecture version. `BIMModelVersion` is an immutable versioned artifact metadata record. `BIMElement` belongs to a model version and carries external identity, type, properties and geometry metadata. `BIMProgressSnapshot` records planned and actual project progress at a date. `BIMImportExport` records controlled exchange jobs.

### BIM runtime
- `BIMImportExport` is the durable queue record for controlled conversion jobs.
- `BIMConverter` processes only explicitly supported adapters; unsupported formats fail visibly rather than being silently approximated.
- Conversion output must become a new immutable `BIMModelVersion` artifact with checksum and manifest before being considered completed by higher-level workflows.
