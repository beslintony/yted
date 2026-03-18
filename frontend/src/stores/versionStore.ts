import { create } from 'zustand';

import { GetVersion } from '../../wailsjs/go/app/App';

interface VersionState {
  version: string;
  isLoading: boolean;
  fetchVersion: () => Promise<void>;
}

export const useVersionStore = create<VersionState>(set => ({
  version: 'dev',
  isLoading: false,
  fetchVersion: async () => {
    set({ isLoading: true });
    try {
      const version = await GetVersion();
      set({ version, isLoading: false });
    } catch (err) {
      console.error('Failed to fetch version:', err);
      set({ isLoading: false });
    }
  },
}));
