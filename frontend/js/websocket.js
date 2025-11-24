class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.connected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        
        // Callbacks
        this.onMessageCallback = null;
        this.onConnectCallback = null;
        this.onDisconnectCallback = null;
        this.onErrorCallback = null;
    }

    // Connect to WebSocket server
    connect() {
        debugLog('Attempting to connect to WebSocket...', this.url);

        try {
            this.ws = new WebSocket(this.url);

            this.ws.onopen = () => {
                debugLog('WebSocket connected!');
                this.connected = true;
                this.reconnectAttempts = 0;
                
                if (this.onConnectCallback) {
                    this.onConnectCallback();
                }
            };

            this.ws.onmessage = (event) => {
                debugLog('WebSocket message received:', event.data);
                
                try {
                    const message = JSON.parse(event.data);
                    if (this.onMessageCallback) {
                        this.onMessageCallback(message);
                    }
                } catch (error) {
                    console.error('Error parsing message:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                if (this.onErrorCallback) {
                    this.onErrorCallback(error);
                }
            };

            this.ws.onclose = () => {
                debugLog('WebSocket disconnected');
                this.connected = false;
                
                if (this.onDisconnectCallback) {
                    this.onDisconnectCallback();
                }

                // Attempt to reconnect
                this.attemptReconnect();
            };

        } catch (error) {
            console.error('Failed to create WebSocket:', error);
        }
    }

    // Attempt to reconnect
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
            
            debugLog(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    // Send a message
    send(message) {
        if (this.connected && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
            debugLog('Message sent:', message);
            return true;
        } else {
            console.error('WebSocket is not connected');
            return false;
        }
    }

    // Disconnect
    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.connected = false;
        }
    }

    // Set callback for incoming messages
    onMessage(callback) {
        this.onMessageCallback = callback;
    }

    // Set callback for connection
    onConnect(callback) {
        this.onConnectCallback = callback;
    }

    // Set callback for disconnection
    onDisconnect(callback) {
        this.onDisconnectCallback = callback;
    }

    // Set callback for errors
    onError(callback) {
        this.onErrorCallback = callback;
    }

    // Check if connected
    isConnected() {
        return this.connected && this.ws && this.ws.readyState === WebSocket.OPEN;
    }
}