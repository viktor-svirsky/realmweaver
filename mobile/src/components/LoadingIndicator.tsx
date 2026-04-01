import React, { useEffect, useState } from 'react';
import { View, Text, StyleSheet } from 'react-native';

const messages = [
  'The dungeon master is thinking...',
  'Rolling dice...',
  'Consulting ancient scrolls...',
  'Weaving the narrative...',
  'The world shifts around you...',
  'Fate is being decided...',
];

export default function LoadingIndicator() {
  const [dots, setDots] = useState('');
  const [msgIdx, setMsgIdx] = useState(() => Math.floor(Math.random() * messages.length));

  useEffect(() => {
    const dotInterval = setInterval(() => {
      setDots((d) => (d.length >= 3 ? '' : d + '.'));
    }, 400);

    const msgInterval = setInterval(() => {
      setMsgIdx((i) => (i + 1) % messages.length);
    }, 4000);

    return () => {
      clearInterval(dotInterval);
      clearInterval(msgInterval);
    };
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.icon}>{'\u{2728}'}</Text>
      <Text style={styles.text}>{messages[msgIdx]}{dots}</Text>
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
