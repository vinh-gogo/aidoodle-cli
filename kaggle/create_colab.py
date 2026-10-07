import json
import os

cells = [
    {
      "cell_type": "markdown",
      "metadata": {},
      "source": [
        "# 1. Tải llama.cpp và Model từ Hugging Face\n",
        "Chạy trên Colab Pro (khuyến nghị chọn A100 hoặc L4 GPU để chạy nhanh và mượt nhất)"
      ]
    },
    {
      "cell_type": "code",
      "metadata": {},
      "execution_count": None,
      "outputs": [],
      "source": [
        "!pip install pyngrok huggingface_hub --quiet\n",
        "\n",
        "# --- THAY ĐỔI MODEL Ở ĐÂY ---\n",
        "HF_REPO = \"mradermacher/Huihui-Qwen3.5-27B-Claude-4.6-Opus-abliterated-i1-GGUF\"\n",
        "MODEL_FILE = \"Huihui-Qwen3.5-27B-Claude-4.6-Opus-abliterated.i1-Q4_K_M.gguf\"\n",
        "# ----------------------------\n",
        "\n",
        "import urllib.request, json, os\n",
        "print('Đang tìm phiên bản llama.cpp CUDA mới nhất...')\n",
        "url = 'https://api.github.com/repos/ggerganov/llama.cpp/releases/latest'\n",
        "res = json.loads(urllib.request.urlopen(url).read().decode())\n",
        "dl_url = next(a['browser_download_url'] for a in res['assets'] if 'ubuntu-cuda-12.' in a['name'] and 'x64.tar.gz' in a['name'] and 'cudart' not in a['name'])\n",
        "print(f'Đang tải {dl_url}...')\n",
        "os.system(f'wget -q -O llama.tar.gz {dl_url} && tar -xzf llama.tar.gz')\n",
        "print('Tải xong llama.cpp!')\n",
        "\n",
        "# Tải model GGUF trực tiếp từ Hugging Face vào thư mục /content/\n",
        "!huggingface-cli download {HF_REPO} {MODEL_FILE} --local-dir /content --local-dir-use-symlinks False\n"
      ]
    },
    {
      "cell_type": "markdown",
      "metadata": {},
      "source": [
        "# 2. Khởi chạy Llama-server ở Background\n",
        "Tối ưu hoá các tham số cho throughput trên GPU xịn (A100/L4)"
      ]
    },
    {
      "cell_type": "code",
      "metadata": {},
      "execution_count": None,
      "outputs": [],
      "source": [
        "import os\n",
        "import subprocess\n",
        "import glob\n",
        "\n",
        "LLAMA_ARG_PORT = \"8000\"\n",
        "env = os.environ.copy()\n",
        "env[\"PYTHONUNBUFFERED\"] = \"1\"  # Fix log streaming\n",
        "\n",
        "# Tìm đường dẫn chính xác của llama-server sau khi giải nén\n",
        "server_path = os.path.abspath(glob.glob('**/llama-server', recursive=True)[0])\n",
        "os.system(f'chmod +x {server_path}')\n",
        "model_path = f'/content/{MODEL_FILE}'\n",
        "\n",
        "command = [\n",
        "    server_path,\n",
        "    \"-m\", model_path,\n",
        "    \"--port\", LLAMA_ARG_PORT,\n",
        "    \n",
        "    # Tham số sinh text\n",
        "    \"--seed\", \"3407\",\n",
        "    \"--temp\", \"0.6\",\n",
        "    \"--top-p\", \"0.95\",\n",
        "    \"--top-k\", \"20\",\n",
        "    \"--min-p\", \"0.0\",\n",
        "    \"--repeat-penalty\", \"1.0\",\n",
        "    \n",
        "    # Tối ưu hoá phần cứng GPU\n",
        "    \"--flash-attn\", \"on\",\n",
        "    \"--kv-unified\",\n",
        "    \"--cache-type-k\", \"q8_0\",\n",
        "    \"--cache-type-v\", \"q8_0\",\n",
        "    \n",
        "    # Tối ưu cho Colab Pro (A100 40GB hoặc L4 24GB)\n",
        "    \"--batch-size\", \"2048\",   # Logical batch size\n",
        "    \"--ubatch-size\", \"512\",   # Physical batch size\n",
        "    \"--n_gpu_layers\", \"99\",   # Offload toàn bộ\n",
        "    \"--ctx-size\", \"32768\",    # Giữ context an toàn 32K\n",
        "    \n",
        "    # Server features\n",
        "    \"--parallel\", \"1\",\n",
        "    \"--cont-batching\",\n",
        "    \"--jinja\",\n",
        "    \"--reasoning\", \"off\",\n",
        "]\n",
        "\n",
        "print(f\"Đang khởi động llama-server tại {server_path} với model {model_path}...\")\n",
        "llama_process = subprocess.Popen(\n",
        "    command,\n",
        "    stdout=subprocess.PIPE,\n",
        "    stderr=subprocess.STDOUT,\n",
        "    text=True,\n",
        "    encoding=\"utf-8\",\n",
        "    errors=\"replace\",\n",
        "    env=env\n",
        ")\n"
      ]
    },
    {
      "cell_type": "markdown",
      "metadata": {},
      "source": [
        "# 3. Mở cổng qua Ngrok\n",
        "Thêm `NGROK_TOKEN` vào tab **Secrets** (Biểu tượng chìa khoá ở thanh công cụ bên trái Colab)."
      ]
    },
    {
      "cell_type": "code",
      "metadata": {},
      "execution_count": None,
      "outputs": [],
      "source": [
        "from pyngrok import ngrok\n",
        "from google.colab import userdata\n",
        "import time\n",
        "import sys\n",
        "\n",
        "try:\n",
        "    ngrok_token = userdata.get(\"NGROK_TOKEN\")\n",
        "    !ngrok config add-authtoken {ngrok_token}\n",
        "except Exception as e:\n",
        "    print(\"LỖI: Vui lòng thêm NGROK_TOKEN vào menu Secrets của Colab!\")\n",
        "\n",
        "ngrok.kill()  # Đóng các tunnel ngrok cũ nếu có để tránh lỗi ERR_NGROK_334\n",
        "public_url = ngrok.connect(LLAMA_ARG_PORT)\n",
        "print(f\"\\n👉 COPY ĐỊA CHỈ NÀY DÁN VÀO CẤU HÌNH API: {public_url}\\n\")\n",
        "\n",
        "# Stream log của server ra output cell\n",
        "for line in llama_process.stdout:\n",
        "    sys.stdout.write(line)\n",
        "    sys.stdout.flush()\n"
      ]
    }
]

notebook = {
    "nbformat": 4,
    "nbformat_minor": 4,
    "metadata": {
        "colab": {
            "provenance": []
        },
        "kernelspec": {
            "name": "python3",
            "display_name": "Python 3",
            "language": "python"
        },
        "language_info": {
            "name": "python"
        }
    },
    "cells": cells
}

with open('kaggle/colab-llama-cpp-ngrok.ipynb', 'w', encoding='utf-8') as f:
    json.dump(notebook, f, ensure_ascii=False, indent=2)
