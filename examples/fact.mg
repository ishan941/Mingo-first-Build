let n = 5;
let acc = 1;
while (n > 1) {
  acc = acc + (acc * (n - 1)) - (acc); // acc *= n; using only +, -, *
  n = n - 1;
}
print(acc);
