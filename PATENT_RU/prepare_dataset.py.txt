import json,sys,hashlib

def main(src,dst):
    n=0; h=hashlib.sha256()
    with open(src,encoding='utf-8') as fi, open(dst,'w',encoding='utf-8') as fo:
        for line in fi:
            if not line.strip(): continue
            x=json.loads(line)
            if 'instruction' in x and 'output' in x:
                item={'messages':[{'role':'user','content':str(x.get('instruction',''))},{'role':'assistant','content':str(x.get('output',''))}]}
            elif 'messages' in x and isinstance(x['messages'],list): item={'messages':x['messages']}
            else: continue
            raw=(json.dumps(item,ensure_ascii=False,separators=(',',':'))+'\n').encode();fo.write(raw.decode());h.update(raw);n+=1
    print(json.dumps({'rows':n,'sha256':h.hexdigest()}))
if __name__=='__main__':
    if len(sys.argv)!=3: raise SystemExit('usage: prepare_dataset.py input.jsonl output.jsonl')
    main(sys.argv[1],sys.argv[2])
