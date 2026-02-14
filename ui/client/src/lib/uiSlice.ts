import { createSlice, type PayloadAction } from '@reduxjs/toolkit';

const SESSION_NAME_STORAGE_KEY = 'vm-system-ui.session-names';
const CURRENT_SESSION_STORAGE_KEY = 'vm-system-ui.current-session-id';

function loadSessionNames(): Record<string, string> {
  try {
    const raw = window.localStorage.getItem(SESSION_NAME_STORAGE_KEY);
    return raw ? (JSON.parse(raw) as Record<string, string>) : {};
  } catch {
    return {};
  }
}

function loadCurrentSessionId(): string | null {
  try {
    return window.localStorage.getItem(CURRENT_SESSION_STORAGE_KEY);
  } catch {
    return null;
  }
}

interface UiState {
  currentSessionId: string | null;
  sessionNames: Record<string, string>;
}

const initialState: UiState = {
  currentSessionId: loadCurrentSessionId(),
  sessionNames: loadSessionNames(),
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setCurrentSessionId(state, action: PayloadAction<string | null>) {
      state.currentSessionId = action.payload;
      // Persist to localStorage
      try {
        if (action.payload) {
          window.localStorage.setItem(CURRENT_SESSION_STORAGE_KEY, action.payload);
        } else {
          window.localStorage.removeItem(CURRENT_SESSION_STORAGE_KEY);
        }
      } catch { /* ignore */ }
    },
    setSessionName(state, action: PayloadAction<{ sessionId: string; name: string }>) {
      const { sessionId, name } = action.payload;
      const trimmed = name.trim();
      if (trimmed) {
        state.sessionNames[sessionId] = trimmed;
      } else {
        delete state.sessionNames[sessionId];
      }
      // Persist to localStorage
      try {
        window.localStorage.setItem(SESSION_NAME_STORAGE_KEY, JSON.stringify(state.sessionNames));
      } catch { /* ignore */ }
    },
  },
});

export const { setCurrentSessionId, setSessionName } = uiSlice.actions;
export default uiSlice.reducer;
