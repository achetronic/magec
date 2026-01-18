/**
 * Centralized error handling with user-friendly messages
 * Errors are logged to console and displayed as notifications
 */

// Error types with user-friendly messages
const ERROR_MESSAGES = {
    // Network errors
    NETWORK_ERROR: 'Connection error. Check your internet connection.',
    TIMEOUT: 'The server took too long to respond. Try again.',
    
    // Server errors
    SERVER_UNAVAILABLE: 'Server is not available. Try again later.',
    SERVER_ERROR: 'Server error occurred. Check console for details.',
    
    // Session errors
    SESSION_CREATE_FAILED: 'Could not create conversation. Check console for details.',
    SESSION_LIST_FAILED: 'Could not load conversations. Check console for details.',
    SESSION_LOAD_FAILED: 'Could not load conversation. Check console for details.',
    SESSION_DELETE_FAILED: 'Could not delete conversation. Check console for details.',
    SESSION_NOT_FOUND: 'Conversation not found.',
    
    // Agent errors
    AGENT_SEND_FAILED: 'Could not send message. Check console for details.',
    AGENT_RESPONSE_FAILED: 'Could not get response. Check console for details.',
    
    // Transcription errors
    TRANSCRIPTION_FAILED: 'Transcription failed. Check console for details.',
    TRANSCRIPTION_SERVER_ERROR: 'Transcription server error. Check console for details.',
    
    // Audio errors
    MICROPHONE_ERROR: 'Could not access microphone. Check permissions.',
    AUDIO_PROCESSING_ERROR: 'Audio processing error. Check console for details.',
    
    // Model loading errors
    MODEL_LOAD_FAILED: 'Could not load model. Check console for details.',
    
    // Generic
    UNKNOWN_ERROR: 'An unexpected error occurred. Check console for details.'
};

// Maps HTTP status codes to error types
const STATUS_CODE_MAP = {
    0: 'NETWORK_ERROR',
    400: 'SERVER_ERROR',
    401: 'SERVER_ERROR',
    403: 'SERVER_ERROR',
    404: 'SESSION_NOT_FOUND',
    408: 'TIMEOUT',
    500: 'SERVER_ERROR',
    502: 'SERVER_UNAVAILABLE',
    503: 'SERVER_UNAVAILABLE',
    504: 'TIMEOUT'
};

/**
 * Error context for categorizing errors
 */
export const ErrorContext = {
    SESSION_CREATE: 'SESSION_CREATE_FAILED',
    SESSION_LIST: 'SESSION_LIST_FAILED',
    SESSION_LOAD: 'SESSION_LOAD_FAILED',
    SESSION_DELETE: 'SESSION_DELETE_FAILED',
    AGENT_SEND: 'AGENT_SEND_FAILED',
    AGENT_RESPONSE: 'AGENT_RESPONSE_FAILED',
    TRANSCRIPTION: 'TRANSCRIPTION_FAILED',
    TRANSCRIPTION_SERVER: 'TRANSCRIPTION_SERVER_ERROR',
    MICROPHONE: 'MICROPHONE_ERROR',
    AUDIO: 'AUDIO_PROCESSING_ERROR',
    MODEL_LOAD: 'MODEL_LOAD_FAILED'
};

/**
 * ErrorHandler - centralized error handling
 */
export class ErrorHandler {
    constructor() {
        this._notificationCallback = null;
    }

    /**
     * Set the callback for showing notifications
     * @param {Function} callback - (type, message) => void
     */
    setNotificationCallback(callback) {
        this._notificationCallback = callback;
    }

    /**
     * Handle an error with context
     * @param {Error|Response} error - The error or failed response
     * @param {string} context - One of ErrorContext values
     * @param {object} options - Additional options
     * @returns {string} User-friendly error message
     */
    handle(error, context, options = {}) {
        const { silent = false, logPrefix = '' } = options;
        
        // Determine the error type
        let errorType = context || 'UNKNOWN_ERROR';
        let technicalDetails = '';
        
        if (error instanceof Response) {
            // HTTP response error
            const statusType = STATUS_CODE_MAP[error.status];
            if (statusType && statusType !== 'SESSION_NOT_FOUND') {
                errorType = statusType;
            }
            technicalDetails = `HTTP ${error.status}`;
        } else if (error instanceof TypeError && error.message.includes('fetch')) {
            // Network error (fetch failed)
            errorType = 'NETWORK_ERROR';
            technicalDetails = error.message;
        } else if (error instanceof Error) {
            technicalDetails = error.message;
        } else if (typeof error === 'string') {
            technicalDetails = error;
        }
        
        // Get user-friendly message
        const userMessage = ERROR_MESSAGES[errorType] || ERROR_MESSAGES.UNKNOWN_ERROR;
        
        // Log to console with full details
        const prefix = logPrefix ? `[${logPrefix}] ` : '';
        console.error(`${prefix}${userMessage}`, {
            context,
            errorType,
            technicalDetails,
            originalError: error
        });
        
        // Show notification
        if (!silent && this._notificationCallback) {
            this._notificationCallback('error', userMessage);
        }
        
        return userMessage;
    }

    /**
     * Wrap a fetch call with error handling
     * @param {Promise<Response>} fetchPromise - The fetch promise
     * @param {string} context - Error context
     * @param {object} options - Options including silent, logPrefix
     * @returns {Promise<Response>} The response or null on error
     */
    async wrapFetch(fetchPromise, context, options = {}) {
        try {
            const response = await fetchPromise;
            if (!response.ok) {
                this.handle(response, context, options);
                return null;
            }
            return response;
        } catch (error) {
            this.handle(error, context, options);
            return null;
        }
    }

    /**
     * Wrap an async operation with error handling
     * @param {Promise} promise - The promise to wrap
     * @param {string} context - Error context
     * @param {object} options - Options
     * @returns {Promise<{data: any, error: string|null}>}
     */
    async wrap(promise, context, options = {}) {
        try {
            const data = await promise;
            return { data, error: null };
        } catch (error) {
            const message = this.handle(error, context, options);
            return { data: null, error: message };
        }
    }
}

// Singleton instance
export const errorHandler = new ErrorHandler();
