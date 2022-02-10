import subprocess

F_NUM = 1
C_NUM = 1

test = """I love music.
I love music.
I love music.
I love music of Kartik.
I love music of Kartik.
Thanks.
I love music of Kartik.
I love music of Kartik."""

test_case_sensitive = """I LOVE MUSIC.
I love music.
I LoVe MuSiC.
I love MuSIC of Kartik.
I love music of kartik.
Thanks.
I love music of kartik.
I love MuSIC of Kartik.
"""

test_first_field_differ = """We love music.
I love music.
They love music.
I love music of Kartik.
We love music of Kartik.
Thanks."""

test_first_letter_differ = """I love music.
A love music.
C love music.
I love music of Kartik.
We love music of Kartik."""

inputs = [
    test,
    test,
    test,
    test,
    test,
    test,
    test_case_sensitive,
    test_first_field_differ,
    test_first_letter_differ
]
args_for_tests = [
    [],
    ["input.txt"],
    ["input.txt", "output.txt"],
    ["-c"],
    ["-d"],
    ["-u"],
    ["-i"],
    ["-f", str(F_NUM)],
    ["-c", str(C_NUM)]
]


def run(args: list, curr_test):
    output = ""

    if "input.txt" in args:
        with open("input.txt", "w") as file:
            file.write(curr_test)

    process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    if "input.txt" not in args:
        process.stdin.write(curr_test.encode())
        process.stdin.close()

    if "output.txt" in args:
        with open("output.txt") as file:
            for line in file:
                output += line
    else:
        for line in process.stdout:
            output += line.decode()
        process.stdout.close()
    process.wait()
    return output


def test_func(args: list, curr_test):
    uniq_output = run(["uniq"] + args, curr_test)
    go_output = run(["go", "run", "main.go"] + args, curr_test)

    print(uniq_output == go_output)


if __name__ == "__main__":
    for i in range(9):
        test_func(args_for_tests[i], inputs[i])

