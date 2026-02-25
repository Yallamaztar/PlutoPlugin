import re

# cleans the log file from ANSI color codes using regex patterns
def main() -> None:
    file = input("enter path to your log file: ")
    with open(file, "r") as f:
        content = f.read()

    with open(file, "w") as f:
        f.write(re.sub(r'\x1B\[[0-9;]*[mK]', '', content))

if __name__ == '__main__':
    main()