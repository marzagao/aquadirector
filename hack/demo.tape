# Aquadirector demo recording
#
# Requirements:
#   brew install vhs
#   brew install ffmpeg
#   # Point vhs at your Chrome installation:
#   export CHROMIUM_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
#
# Generate (run from repo root):
#   CHROMIUM_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" vhs hack/demo.tape
#
# Output: demo.mp4 — add to README with:
#   <video src="demo.mp4" autoplay loop muted playsinline></video>

Output demo.mp4

Set Shell "bash"
Set FontSize 12
Set LineHeight 1.1
Set FontFamily "JetBrains Mono"
Set Width 800
Set Height 700
Set Theme "Dracula"
Set Padding 32
Set TypingSpeed 35ms

# Wire in the mock binary without it being visible
Hide
Type "export PATH=$(pwd)/hack:$PATH"
Enter
Type "clear"
Enter
Show

Sleep 400ms

# --- dashboard ---
Type "aquadirector dashboard"
Sleep 200ms
Enter
Sleep 900ms
Sleep 4s

# --- sensor status as JSON, piped to jq ---
Type "aquadirector sensor status --output json | jq '{ph, temperature_c, sg}'"
Sleep 200ms
Enter
Sleep 600ms
Sleep 2500ms
