# Unary API

The Unary API is a REST API that is used to get the current price of a cryptocurrency.

## Test Specification

describe Unary
  describe NewUnaryAPI
    it should return a new UnaryAPI
  describe GetAssetQuotes
    it should return a list of asset quotes
    when the request fails
      it should return an error
    when there are no symbols set
      it should return an empty list
  describe formatExpiry
    it should return a formatted expiry date
    

