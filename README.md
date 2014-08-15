This is an experiment to measure the overhead of CGO calls. In particular it measures the difference in performance between standard CGO calls, and batched ones (e.g. make one CGO call that calls five C functions).

See discussion at https://github.com/azul3d/issues/issues/17

# Arguments

First, we can measure the cost of pushing arguments onto a stack which could have some copy overhead. Note that in this test CGO calls are still passing zero arguments (only batching is actually passing arguments) so with-CGO timings are not useful here:

Batch Size | Calls | stack-arguments | With CGO | With Batching
-----------|-------|-----------|-----|---------
5 | 350000 | 0 | 45.688143ms | 17.415343ms
5 | 350000 | 5 | 53.212283ms | 26.057713ms
5 | 350000 | 10 | 46.108098ms | 36.445277ms
5 | 350000 | 15 | 46.636866ms | 44.606929ms
5 | 350000 | 20 | 46.52952ms | 51.624302ms
5 | 350000 | 25 | 46.161806ms | 59.803066ms

From which we can determine that:
```
59.80ms - 17.41ms == 42.39ms
42.39ms / 25 arguments == 1.69ms
```
Each argument added also adds an additional 1.69ms overhead because of pushing it onto the stack.

# Calls
Now we can look at the number of calls. A small OpenGL application might use ~2,000 C calls, a large AAA game might use ~100,000. We can see here that as the number of calls increases the CGO overhead, making it more significant:

Batch Size | Calls | stack-arguments | With CGO | With Batching
-----------|-------|-----------|-----|---------
5 | 1000 | 5 | 278.877µs | 151.485µs
5 | 5000 | 5 | 1.31546ms | 743.46µs
5 | 10000 | 5 | 1.298139ms | 786.134µs
5 | 100000 | 5 | 15.337915ms | 7.587906ms
5 | 1000000 | 5 | 132.128883ms | 74.779551ms

This data shows that even with a small number of C calls (1000) batching of only five C calls at a time, can lower the CGO overhead by roughly 50%.

# Batch Size
And one last test, what if we run similar data as above but increase or decrease the batch size (i.e. instead of calling 1 CGO call per five C calls like we did above)

Batch Size | Calls | stack-arguments | With CGO | With Batching
-----------|-------|-----------|-----|---------
1 | 100000 | 5 | 16.047711ms | 21.62803ms
5 | 100000 | 5 | 14.696841ms | 7.61242ms
10 | 100000 | 5 | 13.089094ms | 6.379302ms
15 | 100000 | 5 | 16.557063ms | 5.781183ms
20 | 100000 | 5 | 16.550219ms | 5.548122ms
25 | 100000 | 5 | 14.878079ms | 5.41668ms
30 | 100000 | 5 | 16.301934ms | 5.53583ms

From this we can gather that with a batch size of one (which is identical to not using batching at all) we lose some performance (to be expected). And you can see above that the performance gain is not linear, but at around a batch size of 15 we cap out and stop seeing such performance gains.
