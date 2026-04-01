import React, { useEffect, useState } from 'react';
import { View, Text, Pressable, StyleSheet, FlatList, ScrollView } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../App';
import { rpc } from '../api/nakama';
import { Character } from '../types/game';
import { useGameStore } from '../state/gameStore';

type Props = NativeStackScreenProps<RootStackParamList, 'Home'>;

const languages = [
  { code: 'en', label: 'English', flag: '\u{1F1EC}\u{1F1E7}' },
  { code: 'uk', label: '\u0423\u043A\u0440\u0430\u0457\u043D\u0441\u044C\u043A\u0430', flag: '\u{1F1FA}\u{1F1E6}' },
  { code: 'es', label: 'Espa\u00F1ol', flag: '\u{1F1EA}\u{1F1F8}' },
  { code: 'fr', label: 'Fran\u00E7ais', flag: '\u{1F1EB}\u{1F1F7}' },
  { code: 'de', label: 'Deutsch', flag: '\u{1F1E9}\u{1F1EA}' },
  { code: 'ja', label: '\u65E5\u672C\u8A9E', flag: '\u{1F1EF}\u{1F1F5}' },
  { code: 'ko', label: '\uD55C\uAD6D\uC5B4', flag: '\u{1F1F0}\u{1F1F7}' },
  { code: 'zh', label: '\u4E2D\u6587', flag: '\u{1F1E8}\u{1F1F3}' },
  { code: 'pl', label: 'Polski', flag: '\u{1F1F5}\u{1F1F1}' },
  { code: 'it', label: 'Italiano', flag: '\u{1F1EE}\u{1F1F9}' },
  { code: 'pt', label: 'Portugu\u00EAs', flag: '\u{1F1E7}\u{1F1F7}' },
  { code: 'tr', label: 'T\u00FCrk\u00E7e', flag: '\u{1F1F9}\u{1F1F7}' },
];

export default function HomeScreen({ navigation }: Props) {
  const [characters, setCharacters] = useState<Character[]>([]);
  const language = useGameStore((s) => s.language);
  const setLanguage = useGameStore((s) => s.setLanguage);

  useEffect(() => {
    loadCharacters();
  }, []);

  async function loadCharacters() {
    try {
      const result = await rpc<{ characters: Character[] }>('list_characters', {});
      setCharacters(result.characters || []);
    } catch {
      // No characters yet
    }
  }

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.title}>Realmweaver</Text>
      <Text style={styles.subtitle}>AI Dungeon Master</Text>

      <Text style={styles.sectionTitle}>Language</Text>
      <View style={styles.langGrid}>
        {languages.map((lang) => (
          <Pressable
            key={lang.code}
            style={[
              styles.langButton,
              language === lang.code && styles.langButtonActive,
            ]}
            onPress={() => setLanguage(lang.code)}
          >
            <Text style={styles.langFlag}>{lang.flag}</Text>
            <Text style={[
              styles.langLabel,
              language === lang.code && styles.langLabelActive,
            ]}>{lang.label}</Text>
          </Pressable>
        ))}
      </View>

      <Pressable
        style={styles.newGameButton}
        onPress={() => navigation.navigate('CharacterCreate')}
      >
        <Text style={styles.buttonText}>New Adventure</Text>
      </Pressable>

      {characters.length > 0 && (
        <>
          <Text style={styles.sectionTitle}>Continue</Text>
          <FlatList
            data={characters}
            keyExtractor={(item) => item.id}
            scrollEnabled={false}
            renderItem={({ item }) => (
              <Pressable
                style={styles.characterCard}
                onPress={() => navigation.navigate('Game', { characterId: item.id })}
              >
                <Text style={styles.charName}>{item.name}</Text>
                <Text style={styles.charInfo}>
                  Lv{item.level} {item.class} — HP: {item.hp}/{item.max_hp}
                </Text>
              </Pressable>
            )}
          />
        </>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#16213e',
  },
  content: {
    padding: 24,
    alignItems: 'center',
  },
  title: {
    fontSize: 36,
    fontWeight: 'bold',
    color: '#e0d68a',
    marginTop: 40,
  },
  subtitle: {
    fontSize: 16,
    color: '#a0a0b0',
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    color: '#e0d68a',
    fontWeight: 'bold',
    alignSelf: 'flex-start',
    marginBottom: 10,
    marginTop: 8,
  },
  langGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginBottom: 24,
    justifyContent: 'center',
  },
  langButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1a1a2e',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 20,
    borderWidth: 1,
    borderColor: '#333',
  },
  langButtonActive: {
    borderColor: '#e0d68a',
    backgroundColor: '#2a2a3e',
  },
  langFlag: {
    fontSize: 18,
    marginRight: 6,
  },
  langLabel: {
    color: '#a0a0b0',
    fontSize: 13,
  },
  langLabelActive: {
    color: '#e0d68a',
    fontWeight: 'bold',
  },
  newGameButton: {
    backgroundColor: '#e0d68a',
    paddingHorizontal: 32,
    paddingVertical: 16,
    borderRadius: 8,
    marginBottom: 24,
  },
  buttonText: {
    color: '#1a1a2e',
    fontSize: 18,
    fontWeight: 'bold',
  },
  characterCard: {
    backgroundColor: '#1a1a2e',
    padding: 16,
    borderRadius: 8,
    width: '100%',
    marginBottom: 8,
  },
  charName: {
    color: '#e0d68a',
    fontSize: 18,
    fontWeight: 'bold',
  },
  charInfo: {
    color: '#a0a0b0',
    fontSize: 14,
    marginTop: 4,
  },
});
