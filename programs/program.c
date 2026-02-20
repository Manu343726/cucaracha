// Example C program showing all features supported by cucaracha.
// Function calls, arrays, loops, conditionals, etc. are all demonstrated here.
// This program is designed to run infinitely, so it can be used to test the
// debugger and REPL features.

// Function to compute the nth Fibonacci number
int fib(int n) {
    if (n <= 1) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

// Function to demonstrate array usage
void array_demo() {
    int arr[5];
    for (int i = 0; i < 5; i++) {
        arr[i] = i * i; // Fill array with squares
    }
}

const int MAX = 100;

// Main function
int main() {
    int i = 0;
    while (i < MAX * 2) { // Loop to demonstrate conditionals and function calls
        int result = fib(i);
        i = i + 1;

        if(i > MAX/2) {
            array_demo();
        }

        if(i > MAX) {
            i = 0; // Reset to prevent overflow
        }
    }
    return 0;
}