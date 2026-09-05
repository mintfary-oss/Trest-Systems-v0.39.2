# Stage 3 API examples

Create an object:

```json
{"project_id":"<project-id>","name":"House","address":"Example 1","area_m2":120}
```

Create an estimate item:

```json
{"project_id":"<project-id>","object_id":"<object-id>","name":"Concrete","category":"materials","unit":"m3","quantity":10,"unit_price":8500}
```

Create an order:

```json
{"project_id":"<project-id>","object_id":"<object-id>","title":"Foundation works","amount":120000}
```

Submit a marketplace bid:

```json
{"bid_type":"contractor","amount":110000,"comment":"Can start next week"}
```
