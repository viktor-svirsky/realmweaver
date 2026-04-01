import React, { useState } from 'react';
import { View, TextInput, Pressable, Text, StyleSheet } from 'react-native';
import { sendAction } from '../api/socket';
import { useGameStore } from '../state/gameStore';

export default function PlayerInput() {
  const [text, setText] = useState('');
  const isStreaming = useGameStore((s) => s.isStreaming);

  function handleSend() {
    if (!text.trim() || isStreaming) return;
    sendAction('free_text', { text: text.trim() });
    setText('');
  }

  return (
    <View style={styles.container}>
      <TextInput
        style={styles.input}
        value={text}
        onChangeText={setText}
        placeholder="What do you do?"
        placeholderTextColor="#555"
        onSubmitEditing={handleSend}
        returnKeyType="send"
        editable={!isStreaming}
      />
      <Pressable
        style={[styles.sendButton, (!text.trim() || isStreaming) && styles.disabled]}
        onPress={handleSend}
        disabled={!text.trim() || isStreaming}
      >
        <Text style={styles.sendText}>Go</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingTop: 8,
  },
  input: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    color: '#fff',
    padding: 14,
    borderRadius: 24,
    fontSize: 16,
    borderWidth: 1,
    borderColor: '#333',
  },
  sendButton: {
    backgroundColor: '#e0d68a',
    width: 48,
    height: 48,
    borderRadius: 24,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  disabled: {
    opacity: 0.4,
  },
  sendText: {
    color: '#1a1a2e',
    fontWeight: 'bold',
    fontSize: 16,
  },
});
