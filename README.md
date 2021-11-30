gif画像をcropする  
goroutin使って各フレームで並行処理したい
→ 順番を保証しないといけないから、ただ並行処理するだけじゃだめ
  
## TODO
- [ ] https://stackoverflow.com/questions/37856337/how-to-collect-values-from-n-goroutines-executed-in-a-specific-order みたいに順番を保証する感じに直す
- [ ] エラーハンドリングをチャンネルでやる(errGroupをやめてwaitGroupにする)

## 参考
https://shogo82148.github.io/blog/2015/04/25/quantize-image-in-golang/