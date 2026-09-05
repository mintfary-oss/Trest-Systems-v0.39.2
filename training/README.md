# Trest Training Center

Локальный pipeline для подготовки instruction/ChatML JSONL и обучения LoRA/QLoRA.

## Быстрый запуск

1. Подготовьте `datasets/*.jsonl`.
2. Запустите `python training/scripts/prepare_dataset.py input.jsonl output.jsonl`.
3. На GPU-хосте с `transformers`, `datasets`, `peft`, `trl`, `bitsandbytes` выполните `python training/scripts/train_qlora.py --base-model <local-model> --dataset output.jsonl --output-dir artifacts/<version>`.

Скрипт намеренно не скачивает модель автоматически: базовая модель должна быть локальной/разрешённой и зарегистрированной в Trest. После обучения артефакт регистрируется как новая `ai_model_versions` и проходит evaluation перед promotion.
