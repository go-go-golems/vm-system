import { configureStore } from '@reduxjs/toolkit';
import { vmApi } from './api';
import uiReducer from './uiSlice';

export const store = configureStore({
  reducer: {
    [vmApi.reducerPath]: vmApi.reducer,
    ui: uiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(vmApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
