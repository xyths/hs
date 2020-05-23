# 火币API规则

## 订单更新

订阅订单更新的返回结构，规则如下：

文档地址<https://huobiapi.github.io/docs/spot/v1/cn/#f810bc2ca6>

|`eventType`|`orderStatus`|
|---|---|
|`creation`|`submitted`|
|`trade`|`partial-filled`|
| |`filled`|
|`cancellation`|`partial-canceled`|
| |`canceled`