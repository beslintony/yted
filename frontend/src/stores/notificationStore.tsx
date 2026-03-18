import { create } from 'zustand';

import { Text } from '@mantine/core';
import { modals } from '@mantine/modals';
import { notifications } from '@mantine/notifications';

import type { ConfirmOptions } from '../types';

interface NotificationState {
  // Actions
  showSuccess: (title: string, message?: string) => void;
  showError: (title: string, message?: string) => void;
  showWarning: (title: string, message?: string) => void;
  showInfo: (title: string, message?: string) => void;
  showConfirm: (options: ConfirmOptions) => void;
}

export const useNotificationStore = create<NotificationState>(() => ({
  showSuccess: (title, message) => {
    notifications.show({
      title,
      message: message || '',
      color: 'green',
      icon: '✓',
      autoClose: 3000,
    });
  },

  showError: (title, message) => {
    notifications.show({
      title,
      message: message || '',
      color: 'red',
      icon: '✕',
      autoClose: 5000,
    });
  },

  showWarning: (title, message) => {
    notifications.show({
      title,
      message: message || '',
      color: 'orange',
      icon: '⚠',
      autoClose: 4000,
    });
  },

  showInfo: (title, message) => {
    notifications.show({
      title,
      message: message || '',
      color: 'blue',
      icon: 'ℹ',
      autoClose: 3000,
    });
  },

  showConfirm: options => {
    modals.openConfirmModal({
      title: options.title,
      children: <Text size="sm">{options.message}</Text>,
      labels: {
        confirm: options.confirmLabel || 'Confirm',
        cancel: options.cancelLabel || 'Cancel',
      },
      confirmProps: {
        color: options.confirmColor || 'red',
        variant: 'filled',
      },
      cancelProps: {
        variant: 'light',
      },
      onConfirm: options.onConfirm,
      onCancel: options.onCancel,
      centered: true,
    });
  },
}));

// Hook for easy access to notifications
export const useNotifications = () => {
  const store = useNotificationStore();
  return {
    success: store.showSuccess,
    error: store.showError,
    warning: store.showWarning,
    info: store.showInfo,
    confirm: store.showConfirm,
  };
};
