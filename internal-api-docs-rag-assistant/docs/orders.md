# Orders API

## Create an order

Send `POST /orders` with `customer_id` and a non-empty `items` array.
Each item includes `sku` and a positive `quantity`.
The fictional Atlas orders service returns HTTP 201 with an `order_id` and the state `created`.

## Order cancellation

To cancel an order, send `POST /orders/{order_id}/cancel`.
Cancellation is allowed only while the order is in the `created` state.
A successful cancellation returns HTTP 200 with the state `cancelled`.
An order in the `shipped` state cannot be cancelled and returns HTTP 409 with the code `ORDER_NOT_CANCELLABLE`.
