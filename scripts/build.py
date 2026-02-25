import subprocess
import platform
import sys, os

def get_output() -> str:
    return "plutoplugin.exe" if platform.system().lower() == "windows" else "plutoplugin"
    
def build(cmd: list[str]) -> None:
    try:
        subprocess.check_call(cmd)
    except FileNotFoundError:
        print("Error: Go is not installed or not in PATH")
        sys.exit(1)
    except subprocess.CalledProcessError as e:
        print(f"Build failed with exit code {e.returncode}")
        sys.exit(e.returncode)

    print(f"Build successful")


def main() -> None:
    entry = os.path.join("cmd", "plugin", "main.go")
    if not os.path.exists(entry):
        print(f"Error: file not found {entry}")
        sys.exit(1)

    output = get_output()

    # Go build command & flags
    cmd = [
        "go",
        "build",
        "-ldflags=-s -w -buildid=",
        "-trimpath",
        "-o",
        output,
        entry
    ]

    print(f"Building PlutoPlugin With Flags: ({', '.join([f for f in cmd[1:-2]])})")

    build(cmd)

if __name__ == "__main__":
    main()

if __name__ == '__main__':
    main()