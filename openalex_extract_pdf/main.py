import gzip
import orjson
import os
import pymupdf4llm
from tqdm import tqdm

PDF_DIR = "openalex-pdfs/"
OPENALEX_FILE = "best_works.jsonl.gz"
OUTPUT_FILE = "modified_works.jsonl.gz"

def read_file():
    with gzip.open(OPENALEX_FILE, 'rt') as f:
        for line in f:
            yield orjson.loads(line)

with gzip.open(OUTPUT_FILE, 'wt') as out_f:
    for obj in tqdm(read_file(),total=1500000,desc="processing"):
        title = f"{obj['Id'].split('/')[3]}.pdf"
        file_path = os.path.join(PDF_DIR,title)

        if not os.path.exists(file_path):
            continue

        try:
            text = pymupdf4llm.to_markdown(file_path)
            obj["Text"] = text
            json = orjson.dumps(obj)
            out_f.write(str(json)+"\n")

        except Exception as e:
            continue
