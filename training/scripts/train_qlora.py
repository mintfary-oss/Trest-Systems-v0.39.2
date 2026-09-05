import argparse, json, os

def main():
    p=argparse.ArgumentParser();p.add_argument('--base-model',required=True);p.add_argument('--dataset',required=True);p.add_argument('--output-dir',required=True);p.add_argument('--epochs',type=float,default=1.0);p.add_argument('--batch-size',type=int,default=1);p.add_argument('--grad-accum',type=int,default=8);p.add_argument('--max-seq-length',type=int,default=2048);a=p.parse_args()
    try:
        from datasets import load_dataset
        from transformers import AutoTokenizer, AutoModelForCausalLM, TrainingArguments
        from peft import LoraConfig
        from trl import SFTTrainer
        import torch
    except ImportError as e:
        raise SystemExit('Training dependencies missing: install transformers datasets peft trl bitsandbytes torch on the GPU host') from e
    if not torch.cuda.is_available(): raise SystemExit('CUDA GPU is required for this QLoRA runner')
    os.makedirs(a.output_dir,exist_ok=True)
    ds=load_dataset('json',data_files=a.dataset,split='train')
    tok=AutoTokenizer.from_pretrained(a.base_model,local_files_only=True)
    model=AutoModelForCausalLM.from_pretrained(a.base_model,local_files_only=True,device_map='auto',torch_dtype='auto')
    if tok.pad_token is None: tok.pad_token=tok.eos_token
    def fmt(x): return tok.apply_chat_template(x['messages'],tokenize=False,add_generation_prompt=False)
    cfg=LoraConfig(r=16,lora_alpha=32,lora_dropout=0.05,target_modules='all-linear',task_type='CAUSAL_LM')
    args=TrainingArguments(output_dir=a.output_dir,num_train_epochs=a.epochs,per_device_train_batch_size=a.batch_size,gradient_accumulation_steps=a.grad_accum,logging_steps=10,save_strategy='epoch',bf16=torch.cuda.is_bf16_supported(),fp16=not torch.cuda.is_bf16_supported(),report_to=[])
    trainer=SFTTrainer(model=model,tokenizer=tok,train_dataset=ds,formatting_func=fmt,peft_config=cfg,args=args,max_seq_length=a.max_seq_length)
    trainer.train();trainer.save_model(a.output_dir);tok.save_pretrained(a.output_dir)
    with open(os.path.join(a.output_dir,'training_manifest.json'),'w',encoding='utf-8') as f: json.dump({'base_model':a.base_model,'dataset':a.dataset,'method':'qlora','epochs':a.epochs},f,ensure_ascii=False,indent=2)
if __name__=='__main__': main()
