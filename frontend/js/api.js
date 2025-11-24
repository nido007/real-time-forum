class APIClient {
  constructor(baseURL) {
      this.baseURL = baseURL;
  }

  // Generic fetch wrapper
  async request(endpoint, options = {}) {
      const url = `${this.baseURL}${endpoint}`;
      
      const defaultOptions = {
          headers: {
              'Content-Type': 'application/json',
          },
          credentials: 'include', // Include cookies for session
      };

      const config = { ...defaultOptions, ...options };

      try {
          debugLog(`API Request: ${config.method || 'GET'} ${url}`);
          const response = await fetch(url, config);
          
          if (!response.ok) {
              throw new Error(`HTTP error! status: ${response.status}`);
          }

          const data = await response.json();
          debugLog('API Response:', data);
          return data;

      } catch (error) {
          console.error('API Error:', error);
          throw error;
      }
  }

  // Get online users
  async getOnlineUsers() {
      return this.request('/online-users');
  }

  // Get message history with a user
  async getMessageHistory(userID) {
      return this.request(`/messages/history?user_id=${userID}`);
  }

  // Mark messages as read
  async markMessagesAsRead(userID) {
      return this.request(`/messages/read`, {
          method: 'POST',
          body: JSON.stringify({ user_id: userID })
      });
  }

  // Get current user info
  async getCurrentUser() {
      return this.request('/me');
  }
}