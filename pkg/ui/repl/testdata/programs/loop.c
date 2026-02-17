// loop.c - Program with loops for testing breakpoints
int sum_to_n(int n) {
    int sum = 0;
    for (int i = 1; i <= n; i++) {
        sum += i;
    }
    return sum;
}

int main() {
    int result = sum_to_n(10);
    return result;
}
