// Configuration for the application
const CONFIG = {
  // WebSocket URL
  WS_URL: 'ws://localhost:8080/ws',
  
  // API Base URL
  API_URL: 'http://localhost:8080/api',
  
  // Current user (will be set after login)
  currentUser: null,
  
  // Debug mode
  DEBUG: true
};

// Helper function for logging
function debugLog(message, data) {
  if (CONFIG.DEBUG) {
      console.log(`[DEBUG] ${message}`, data || '');
  }
}