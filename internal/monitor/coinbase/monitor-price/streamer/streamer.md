
## test specs
describe Streamer
	describe NewStreamer
		it should return a new Streamer
	describe Start
    before each
      reset the websocket mock
      start a mock websocket server
    after each
      stop the mock websocket server
		it should start the streamer without an error
    when the streamer is already started
      it should return the error "streamer already started"
    when the url is not set
      it should not start the streamer
    when the context is cancelled while trying to connect to the websocket
      it should return an error containing the text "connection aborted"
    describe readStreamQuote
      when a price tick is received
        it should send the price tick to the channel
      when an extended quote is received
        it should send the extended quote to the channel
      when a message is not a price tick or extended quote
        it should not send anything to the channel
      when there is an error reading from the stream
        it should stop listening for messages
        [pending] it should return the error
    describe writeStreamSubscription
      it should write the subscription to the stream
      when there is an error writing to the stream
        it should stop listening for messages
        [pending] it should return the error
  describe SetSymbolsAndUpdateSubscriptions
    before each
      reset the websocket mock
      start a mock websocket server
    after each
      stop the mock websocket server
    it should set the symbols and not return an error
    it should subscribe to the stream for the new symbols
    when there is an existing subscription
      it should unsubscribe from the stream for the old symbols
    when the streamer is not started
      it should return early without error
    when there is an error subscribing to the stream
      [pending] it should return the error
    when there is an error unsubscribing from the stream
      [pending] it should return the error
  describe SetURL
    it should set the url and not return an error
    when the streamer is started
      it should return the error "cannot set URL while streamer is connected"
