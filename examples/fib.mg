let n = 8;
let a = 0;
let b = 1;
let i = 0;
while (i < n) {
  let next = a + b;
  a = b;
  b = next;
  i = i + 1;
}
print(a);
