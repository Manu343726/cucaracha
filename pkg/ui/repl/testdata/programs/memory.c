// memory.c - Program that uses arrays for memory inspection
int main() {
    int array[5] = {10, 20, 30, 40, 50};
    int sum = 0;
    for (int i = 0; i < 5; i++) {
        sum += array[i];
    }
    return sum;
}
