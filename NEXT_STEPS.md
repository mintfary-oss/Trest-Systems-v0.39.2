# Release acceptance not yet evidenced

1. Run this exact ZIP on a disposable Ubuntu/Debian host with Docker, record install + repeat install + reboot.
2. Exercise login/catalog/order/project/BIM/RAG flows in a browser. Do not equate HTTP 200 with full functionality.
3. Review conversion of the old public-schema marketplace on a restored copy before updating production.
4. Back up/restore SQL and object/file volumes; archive images/models via the included exporter.
5. Rebuild with a maintained Go toolchain and complete a current dependency/security audit.
