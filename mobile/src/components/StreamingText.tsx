import React from 'react';
import { Text, StyleSheet, View } from 'react-native';

interface Props {
  text: string;
}

export default function StreamingText({ text }: Props) {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>
        {text}
        <Text style={styles.cursor}>|</Text>
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 4,
    paddingBottom: 8,
  },
  text: {
    color: '#d4d4d4',
    fontSize: 16,
    lineHeight: 24,
  },
  cursor: {
    color: '#e0d68a',
    fontWeight: 'bold',
  },
});
