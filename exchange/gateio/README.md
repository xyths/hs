# gate.io API 封装说明

此文档中，记录了gate接口的一些重要细节和封装后的使用注意事项。

# V4版本API

V4版本分为
- spot: 现货
- margin: 杠杆
- future: 永续合约
- delivery: 交割合约

### `order.query`

返回的数据如下:

```json
{
  "id": 21953668381,
  "market": "SERO_USDT",
  "tif": 1,
  "user": 2931605,
  "ctime": 1608694438.202096,
  "mtime": 1608694438.202096,
  "price": "0.1062",
  "amount": "235.627",
  "iceberg": "0",
  "left": "235.627",
  "deal_fee_rebate": "0",
  "deal_point_fee": "0",
  "gt_discount": "1",
  "gt_taker_fee": "0.00045",
  "gt_maker_fee": "0.00015",
  "deal_gt_fee": "0",
  "orderType": 1,
  "type": 1,
  "dealFee": "0",
  "filledAmount": "0",
  "filledTotal": "0"
}
```