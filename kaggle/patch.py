import json
import os

with open('kaggle/colab-llama-cpp-ngrok.ipynb', 'r', encoding='utf-8') as f:
    nb = json.load(f)

nb['cells'][1]['source'] = [
    '!pip install -U "huggingface_hub[cli]" pyngrok --quiet\n',
    '\n',
    '# --- THAY ĐỔI MODEL Ở ĐÂY ---\n',
    'HF_REPO = "mradermacher/Huihui-Qwen3.5-27B-Claude-4.6-Opus-abliterated-i1-GGUF"\n',
    'MODEL_FILE = "Huihui-Qwen3.5-27B-Claude-4.6-Opus-abliterated.i1-Q4_K_M.gguf"\n',
    '# ----------------------------\n',
    '\n',
    '# Clone và build llama.cpp hỗ trợ CUDA\n',
    '!git clone https://github.com/ggerganov/llama.cpp\n',
    '!cd llama.cpp && make GGML_CUDA=1 -j\n',
    '\n',
    '# Tải model GGUF trực tiếp từ Hugging Face\n',
    '!huggingface-cli download {HF_REPO} {MODEL_FILE} --local-dir . --local-dir-use-symlinks False\n'
]

nb['cells'][3]['source'] = [
    'import os\n',
    'import subprocess\n',
    '\n',
    'LLAMA_ARG_PORT = "8000"\n',
    'env = os.environ.copy()\n',
    '\n',
    'command = [\n',
    '    "./llama.cpp/llama-server",\n',
    '    "-m", MODEL_FILE,  # Dùng biến từ Cell trước\n',
    '    "--port", LLAMA_ARG_PORT,\n',
    '    \n',
    '    # Tham số sinh text\n',
    '    "--seed", "3407",\n',
    '    "--temp", "0.6",\n',
    '    "--top-p", "0.95",\n',
    '    "--top-k", "20",\n',
    '    "--min-p", "0.0",\n',
    '    "--repeat-penalty", "1.0",\n',
    '    \n',
    '    # Tối ưu hoá phần cứng GPU\n',
    '    "--flash-attn", "on",\n',
    '    "--kv-unified",\n',
    '    "--cache-type-k", "q8_0",\n',
    '    "--cache-type-v", "q8_0",\n',
    '    \n',
    '    # Tối ưu cho Colab Pro (A100 40GB hoặc L4 24GB)\n',
    '    "--batch-size", "2048",   # Logical batch size\n',
    '    "--ubatch-size", "512",   # Physical batch size\n',
    '    "--n_gpu_layers", "99",   # Offload toàn bộ\n',
    '    "--ctx-size", "32768",    # Giữ context an toàn 32K\n',
    '    \n',
    '    # Server features\n',
    '    "--parallel", "1",\n',
    '    "--cont-batching",\n',
    '    "--jinja",\n',
    '    "--chat-template-kwargs", \'{"enable_thinking":false}\',\n',
    ']\n',
    '\n',
    'print(f"Đang khởi động llama-server với model {MODEL_FILE}...")\n',
    'llama_process = subprocess.Popen(\n',
    '    command,\n',
    '    stdout=subprocess.PIPE,\n',
    '    stderr=subprocess.STDOUT,\n',
    '    text=True,\n',
    '    encoding="utf-8",\n',
    '    errors="replace",\n',
    '    env=env\n',
    ')\n'
]

with open('kaggle/colab-llama-cpp-ngrok.ipynb', 'w', encoding='utf-8') as f:
    json.dump(nb, f, ensure_ascii=False, indent=2)
