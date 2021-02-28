# Profit 💰
Store cryptocurrency wallet balance and keep track of changing profit over time. 

## Technologies
- Go (1.13.5)
- DynamoDB
- AWS Lambda

## Installation
Clone the repository \
```git clone https://github.com/berat703/profit && cd profit```\

Create .env file to set environment variables. \
```cp .env-example .env```

## Deploy to AWS Lambda
I'm using a tool that helps to deploy serverless application with several basic commands.\ I'm not going to tell how it works.It's beyond of our subject. https://www.serverless.com/framework/docs/getting-started/ Before you deploy the application to Lambda, don't forget to change credentials in .env file\

``make``\
``serverless deploy``

## Endpoints

Currently there's just one endpoint and it's store current wallet status to DynamoDB when sending GET request.\

### [GET] /

It's responding current wallet status and store it to DynamoDB.

An example about success response

```
{
  "id": "b1c0d9e1-3d95-45f4-b4ae-c3193024e77b",
  "coins": [
    {
      "asset": "BNB",
      "balance": 0.04928775876760483,
      "total_as_btc": 0.00022701942416634285
    },
    {
      "asset": "ETH",
      "balance": 8.601390381460078,
      "total_as_btc": 0.001974879249283608
    }
    ...
  ],
  "balance_as_usd": 1710.871848304593,
  "balance_as_btc": 0.02867812331156098,
  "btc_usd": 45158.4609375,
  "created_at": 1614553066
}
```

## TODO 
- CRON (It should run per day)
- Telegram Bot (To check daily, weekly, monthly profit)
