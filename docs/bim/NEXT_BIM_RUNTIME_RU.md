# Следующий BIM runtime

1. Полноценный IFC semantic parser/serializer и mapping IFC element -> `bim_elements`.
2. DXF reader/writer через существующий модуль proektirovka с сохранением GOST-геометрии.
3. OBJ export из BIM element graph.
4. Object-storage workflow для больших моделей и потоковых обменов.
5. Viewer: element tree, selection, properties panel и подсветка changed/added/removed geometry.
6. Planned/actual progress overlays и связь с BIM element quantities.
7. Job retry/backoff policy + cancel audit event + output validation.
8. Project membership/ownership checks для всех BIM endpoints.
