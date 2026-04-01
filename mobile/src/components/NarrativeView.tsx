import React, { useRef, useEffect } from 'react';
import { FlatList, Text, View, StyleSheet, Platform } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { NarrativeEntry } from '../types/game';
import LoadingIndicator from './LoadingIndicator';

// Simple markdown renderer for **bold** and *italic*
function MarkdownText({ text, style }: { text: string; style: object }) {
  const parts: React.ReactNode[] = [];
  // Match **bold**, *italic*, and plain text
  const regex = /(\*\*[^*]+\*\*|\*[^*]+\*)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;
  let key = 0;

  while ((match = regex.exec(text)) !== null) {
    // Add text before match
    if (match.index > lastIndex) {
      parts.push(
        <Text key={key++} style={style}>{text.slice(lastIndex, match.index)}</Text>
      );
    }

    const matched = match[0];
    if (matched.startsWith('**')) {
      parts.push(
        <Text key={key++} style={[style, styles.bold]}>{matched.slice(2, -2)}</Text>
      );
    } else if (matched.startsWith('*')) {
      parts.push(
        <Text key={key++} style={[style, styles.italic]}>{matched.slice(1, -1)}</Text>
      );
    }
    lastIndex = match.index + matched.length;
  }

  // Add remaining text
  if (lastIndex < text.length) {
    parts.push(
      <Text key={key++} style={style}>{text.slice(lastIndex)}</Text>
    );
  }

  return <Text>{parts}</Text>;
}

export default function NarrativeView() {
  const entries = useGameStore((s) => s.narrativeEntries);
  const isLoading = useGameStore((s) => s.isLoading);
  const listRef = useRef<FlatList>(null);

  useEffect(() => {
    if (listRef.current && entries.length > 0) {
      setTimeout(() => listRef.current?.scrollToEnd({ animated: true }), 100);
    }
  }, [entries.length, isLoading]);

  const renderEntry = ({ item, index }: { item: NarrativeEntry; index: number }) => {
    const showSeparator = index > 0;

    return (
      <View>
        {showSeparator && <View style={styles.separator} />}
        {item.type === 'narrative' ? (
          <MarkdownText text={item.text} style={styles.narrative} />
        ) : item.type === 'mechanical' ? (
          <View style={styles.mechanicalBox}>
            <Text style={styles.mechanicalIcon}>{'\u{2694}\uFE0F'}</Text>
            <Text style={styles.mechanical}>{item.text}</Text>
          </View>
        ) : item.text.startsWith('>') ? (
          <View style={styles.playerActionBox}>
            <Text style={styles.playerAction}>{item.text.substring(2)}</Text>
          </View>
        ) : (
          <View style={styles.systemBox}>
            <Text style={styles.system}>{item.text}</Text>
          </View>
        )}
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <FlatList
        ref={listRef}
        data={entries}
        keyExtractor={(item) => item.id}
        renderItem={renderEntry}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
      />
      {isLoading && <LoadingIndicator />}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingHorizontal: 16,
  },
  listContent: {
    paddingVertical: 8,
    paddingBottom: 16,
  },
  narrative: {
    color: '#d4d4d4',
    fontSize: 16,
    lineHeight: 26,
  },
  bold: {
    fontWeight: 'bold',
    color: '#e0d68a',
  },
  italic: {
    fontStyle: 'italic',
    color: '#b8c0cc',
  },
  separator: {
    height: 1,
    backgroundColor: '#ffffff15',
    marginVertical: 14,
  },
  mechanicalBox: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(224, 214, 138, 0.08)',
    padding: 10,
    borderRadius: 8,
    borderLeftWidth: 3,
    borderLeftColor: '#e0d68a',
    marginVertical: 6,
    gap: 8,
  },
  mechanicalIcon: {
    fontSize: 14,
  },
  mechanical: {
    color: '#e0d68a',
    fontSize: 13,
    fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }),
    flex: 1,
  },
  playerActionBox: {
    backgroundColor: 'rgba(46, 204, 113, 0.08)',
    paddingVertical: 8,
    paddingHorizontal: 14,
    borderRadius: 12,
    borderLeftWidth: 3,
    borderLeftColor: '#2ecc71',
    alignSelf: 'flex-end',
    maxWidth: '80%',
    marginVertical: 4,
  },
  playerAction: {
    color: '#2ecc71',
    fontSize: 14,
    fontWeight: 'bold',
    fontStyle: 'italic',
  },
  systemBox: {
    backgroundColor: 'rgba(93, 173, 226, 0.1)',
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 20,
    alignSelf: 'center',
    marginVertical: 8,
  },
  system: {
    color: '#5dade2',
    fontSize: 14,
    fontWeight: 'bold',
    textAlign: 'center',
  },
});
