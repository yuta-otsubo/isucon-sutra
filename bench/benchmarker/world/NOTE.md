#### bench/benchmarker/world/chair.go
- sync/atomic とは？
    - 複数のゴルーチンから同時にアクセスされる変数を安全に読み書きする
    - mutex に比べて高速で、単一の値を安全に操作したい時に使用する
