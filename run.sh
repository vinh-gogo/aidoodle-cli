#!/usr/bin/env bash
# setup-ainovel.sh
# Tự động: tải model Qwen qua Ollama, tạo bản có context lớn, ghi ~/.ainovel/config.json
# Chạy được trong Git Bash (Windows), WSL, macOS, Linux.

set -euo pipefail

# ---------- Mặc định (có thể đổi bằng tham số) ----------
BASE_MODEL="qwen3.5:4b"   # -m
NUM_CTX=16384             # -c
REASONING="off"           # -r : off|low|medium|high|xhigh|max
SKIP_MODEL=0              # -s : bỏ qua bước Ollama, chỉ ghi config
SET_ENV=0                 # -e : đặt biến môi trường tiết kiệm VRAM (Windows: setx)
FORCE=0                   # -f : ghi đè config cũ không hỏi (vẫn tự sao lưu)
OLLAMA_URL="http://localhost:11434"

usage() {
  cat <<EOF
Cách dùng: bash setup-ainovel.sh [tùy chọn]

  -m <model>   Model gốc trên Ollama      (mặc định: ${BASE_MODEL})
  -c <số>      Kích thước context num_ctx (mặc định: ${NUM_CTX})
  -r <mức>     reasoning_effort           (mặc định: ${REASONING})
  -s           Bỏ qua Ollama, chỉ ghi file config
  -e           Đặt OLLAMA_FLASH_ATTENTION và OLLAMA_KV_CACHE_TYPE (tùy chọn)
  -f           Ghi đè config cũ không hỏi (vẫn tạo bản sao lưu)
  -h           Hiện trợ giúp

Ví dụ:
  bash setup-ainovel.sh
  bash setup-ainovel.sh -c 8192
  bash setup-ainovel.sh -m qwen3.5:9b -c 8192 -e
EOF
}

die()  { echo "LỖI: $*" >&2; exit 1; }
info() { echo "==> $*"; }

while getopts ":m:c:r:sefh" opt; do
  case "$opt" in
    m) BASE_MODEL="$OPTARG" ;;
    c) NUM_CTX="$OPTARG" ;;
    r) REASONING="$OPTARG" ;;
    s) SKIP_MODEL=1 ;;
    e) SET_ENV=1 ;;
    f) FORCE=1 ;;
    h) usage; exit 0 ;;
    :) die "Tùy chọn -$OPTARG cần một giá trị. Xem: bash $0 -h" ;;
    \?) die "Tùy chọn không hợp lệ: -$OPTARG. Xem: bash $0 -h" ;;
  esac
done

[[ "$NUM_CTX" =~ ^[0-9]+$ ]] && (( NUM_CTX >= 2048 )) || die "num_ctx phải là số nguyên >= 2048 (đang là: $NUM_CTX)"
case "$REASONING" in
  off|low|medium|high|xhigh|max) ;;
  *) die "reasoning_effort phải là một trong: off low medium high xhigh max" ;;
esac

# Tên model tùy biến, ví dụ qwen3.5:4b + 16384 -> qwen3.5-4b-16k
CUSTOM_MODEL="${BASE_MODEL//:/-}-$((NUM_CTX / 1024))k"

# ---------- Thư mục home (Git Bash: ưu tiên USERPROFILE của Windows) ----------
if [[ -n "${USERPROFILE:-}" ]] && command -v cygpath >/dev/null 2>&1; then
  HOME_DIR="$(cygpath -u "$USERPROFILE")"
else
  HOME_DIR="$HOME"
fi
CONFIG_DIR="$HOME_DIR/.ainovel"
CONFIG_FILE="$CONFIG_DIR/config.json"

# ---------- Bước 1: Ollama ----------
if [[ $SKIP_MODEL -eq 0 ]]; then
  command -v ollama >/dev/null 2>&1 \
    || die "Không thấy lệnh 'ollama'. Cài Ollama trước, rồi mở lại Git Bash."
  command -v curl >/dev/null 2>&1 || die "Không thấy lệnh 'curl'."

  if ! curl -fsS --max-time 5 "$OLLAMA_URL/" >/dev/null 2>&1; then
    die "Ollama server chưa chạy tại $OLLAMA_URL. Hãy mở ứng dụng Ollama (hoặc chạy 'ollama serve' ở cửa sổ khác) rồi chạy lại."
  fi

  info "Tải model gốc: $BASE_MODEL"
  ollama pull "$BASE_MODEL"

  info "Tạo model $CUSTOM_MODEL (num_ctx=$NUM_CTX)"
  MODELFILE="./.Modelfile.ainovel.$$"
  trap 'rm -f "$MODELFILE"' EXIT
  printf 'FROM %s\nPARAMETER num_ctx %s\n' "$BASE_MODEL" "$NUM_CTX" > "$MODELFILE"
  ollama create "$CUSTOM_MODEL" -f "$MODELFILE"
else
  info "Bỏ qua bước Ollama (-s). Giả định model '$CUSTOM_MODEL' đã được tạo."
fi

# ---------- Bước 2: biến môi trường tiết kiệm VRAM (tùy chọn) ----------
if [[ $SET_ENV -eq 1 ]]; then
  if command -v setx >/dev/null 2>&1; then
    info "Đặt biến môi trường cho Ollama (Windows)"
    setx OLLAMA_FLASH_ATTENTION 1 >/dev/null
    setx OLLAMA_KV_CACHE_TYPE q8_0 >/dev/null
    echo "    Hãy thoát Ollama ở khay hệ thống rồi mở lại để áp dụng."
  else
    echo "    Không phải Windows. Hãy tự thêm vào shell profile hoặc service của Ollama:"
    echo "      export OLLAMA_FLASH_ATTENTION=1"
    echo "      export OLLAMA_KV_CACHE_TYPE=q8_0"
  fi
fi

# ---------- Bước 3: ghi config ----------
mkdir -p "$CONFIG_DIR"

if [[ -f "$CONFIG_FILE" ]]; then
  if [[ $FORCE -eq 0 ]]; then
    read -r -p "Đã có $CONFIG_FILE. Ghi đè (sẽ tự sao lưu)? [y/N] " ans || ans=""
    [[ "$ans" =~ ^[Yy]$ ]] || die "Đã hủy, không thay đổi config. Dùng -f để ghi đè không hỏi."
  fi
  BACKUP="$CONFIG_FILE.bak.$(date +%Y%m%d-%H%M%S)"
  cp "$CONFIG_FILE" "$BACKUP"
  info "Đã sao lưu config cũ: $BACKUP"
fi

cat > "$CONFIG_FILE" <<EOF
{
  "provider": "ollama",
  "model": "${CUSTOM_MODEL}",
  "reasoning_effort": "${REASONING}",
  "providers": {
    "ollama": {
      "base_url": "${OLLAMA_URL}/v1",
      "models": [
        { "name": "${CUSTOM_MODEL}", "context_window": ${NUM_CTX} }
      ]
    }
  },
  "style": "doodle-explainer"
}
EOF

info "Đã ghi config: $CONFIG_FILE"
echo
echo "Xong. Các bước tiếp theo:"
echo "  1. Kiểm tra model dùng 100% GPU:  ollama run $CUSTOM_MODEL \"xin chào\"   rồi  ollama ps"
echo "     (nếu bị chia CPU/GPU, chạy lại với -c 8192)"
echo "  2. Tạo thư mục riêng cho truyện:  mkdir truyen1 && cd truyen1"
echo "  3. Chạy:                          ainovel-cli"