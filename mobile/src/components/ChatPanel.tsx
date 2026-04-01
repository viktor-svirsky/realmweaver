import React, { useEffect, useState, useRef } from 'react';
import { View, Text, TextInput, Pressable, FlatList, StyleSheet } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { getChat, postChat, ChatMessage } from '../api/social';
import { t } from '../i18n/translations';

interface Props {
  onClose: () => void;
}

export default function ChatPanel({ onClose }: Props) {
  const character = useGameStore((s) => s.character);
  const language = useGameStore((s) => s.language);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [text, setText] = useState('');
  const [sending, setSending] = useState(false);
  const listRef = useRef<FlatList>(null);

  const regionX = character?.region_x ?? 0;
  const regionY = character?.region_y ?? 0;

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const msgs = await getChat(regionX, regionY);
        if (!cancelled) setMessages(msgs);
      } catch {
        // ignore
      }
    }
    load();
    const interval = setInterval(load, 5000);
    return () => { cancelled = true; clearInterval(interval); };
  }, [regionX, regionY]);

  async function handleSend() {
    if (!text.trim() || !character || sending) return;
    setSending(true);
    try {
      await postChat(regionX, regionY, character.name, text.trim());
      setText('');
      // Refresh messages
      const msgs = await getChat(regionX, regionY);
      setMessages(msgs);
    } catch {
      // ignore
    }
    setSending(false);
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>{'\u{1F4AC}'} {t('region_chat', language)}</Text>
        <Pressable style={styles.closeBtn} onPress={onClose}>
          <Text style={styles.closeText}>{t('close', language)}</Text>
        </Pressable>
      </View>

      <FlatList
        ref={listRef}
        data={messages}
        keyExtractor={(item) => String(item.id)}
        renderItem={({ item }) => (
          <View style={styles.msgRow}>
            <Text style={styles.msgText}>{item.content}</Text>
            <Text style={styles.msgTime}>{item.created_at?.split('T')[1]?.split('.')[0] || ''}</Text>
          </View>
        )}
        contentContainerStyle={styles.msgList}
        onContentSizeChange={() => listRef.current?.scrollToEnd({ animated: false })}
        ListEmptyComponent={
          <Text style={styles.emptyText}>{t('no_messages', language)}</Text>
        }
      />

      <View style={styles.inputRow}>
        <TextInput
          style={styles.input}
          value={text}
          onChangeText={setText}
          placeholder={t('say_something', language)}
          placeholderTextColor="#555"
          maxLength={500}
          onSubmitEditing={handleSend}
          returnKeyType="send"
        />
        <Pressable
          style={[styles.sendBtn, (!text.trim() || sending) && styles.disabled]}
          onPress={handleSend}
          disabled={!text.trim() || sending}
        >
          <Text style={styles.sendText}>{t('send', language)}</Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: '50%',
    backgroundColor: '#0d1117',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    borderTopWidth: 1,
    borderTopColor: '#2a2a3e',
    zIndex: 50,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#1a1a2e',
  },
  title: {
    color: '#e0d68a',
    fontSize: 16,
    fontWeight: 'bold',
  },
  closeBtn: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 16,
  },
  closeText: {
    color: '#e0d68a',
    fontSize: 12,
    fontWeight: 'bold',
  },
  msgList: {
    padding: 12,
    paddingBottom: 4,
  },
  msgRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 8,
    gap: 8,
  },
  msgText: {
    color: '#d4d4d4',
    fontSize: 14,
    flex: 1,
  },
  msgTime: {
    color: '#444',
    fontSize: 10,
  },
  emptyText: {
    color: '#444',
    fontSize: 14,
    textAlign: 'center',
    marginTop: 40,
  },
  inputRow: {
    flexDirection: 'row',
    padding: 10,
    gap: 8,
    borderTopWidth: 1,
    borderTopColor: '#1a1a2e',
  },
  input: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    color: '#fff',
    padding: 10,
    borderRadius: 20,
    fontSize: 14,
  },
  sendBtn: {
    backgroundColor: '#e0d68a',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
    justifyContent: 'center',
  },
  disabled: {
    opacity: 0.4,
  },
  sendText: {
    color: '#1a1a2e',
    fontWeight: 'bold',
    fontSize: 14,
  },
});
