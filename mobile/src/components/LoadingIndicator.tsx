import React, { useEffect, useState } from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { t } from '../i18n/translations';

const messageKeys = [
  'loading_1',
  'loading_2',
  'loading_3',
  'loading_4',
  'loading_5',
  'loading_6',
];

export default function LoadingIndicator() {
  const language = useGameStore((s) => s.language);
  const [dots, setDots] = useState('');
  const [msgIdx, setMsgIdx] = useState(() => Math.floor(Math.random() * messageKeys.length));

  useEffect(() => {
    const dotInterval = setInterval(() => {
      setDots((d) => (d.length >= 3 ? '' : d + '.'));
    }, 400);

    const msgInterval = setInterval(() => {
      setMsgIdx((i) => (i + 1) % messageKeys.length);
    }, 4000);

    return () => {
      clearInterval(dotInterval);
      clearInterval(msgInterval);
    };
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.icon}>{'\u{2728}'}</Text>
      <Text style={styles.text}>{t(messageKeys[msgIdx], language)}{dots}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 8,
  },
  icon: {
    fontSize: 18,
  },
  text: {
    color: '#e0d68a',
    fontSize: 14,
    fontStyle: 'italic',
    opacity: 0.8,
  },
});
