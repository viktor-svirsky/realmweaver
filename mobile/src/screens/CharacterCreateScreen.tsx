import React, { useState } from 'react';
import { View, Text, TextInput, Pressable, StyleSheet, ScrollView } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../App';
import { rpc } from '../api/nakama';
import { Class, Character } from '../types/game';
import { useGameStore } from '../state/gameStore';
import { t } from '../i18n/translations';

type Props = NativeStackScreenProps<RootStackParamList, 'CharacterCreate'>;

const classes: { id: Class; nameKey: string; descKey: string; color: string }[] = [
  { id: 'warrior', nameKey: 'class_warrior', descKey: 'class_warrior_desc', color: '#c0392b' },
  { id: 'mage', nameKey: 'class_mage', descKey: 'class_mage_desc', color: '#2980b9' },
  { id: 'rogue', nameKey: 'class_rogue', descKey: 'class_rogue_desc', color: '#27ae60' },
  { id: 'cleric', nameKey: 'class_cleric', descKey: 'class_cleric_desc', color: '#f39c12' },
  { id: 'ranger', nameKey: 'class_ranger', descKey: 'class_ranger_desc', color: '#2ecc71' },
  { id: 'paladin', nameKey: 'class_paladin', descKey: 'class_paladin_desc', color: '#f1c40f' },
  { id: 'necromancer', nameKey: 'class_necromancer', descKey: 'class_necromancer_desc', color: '#8e44ad' },
  { id: 'berserker', nameKey: 'class_berserker', descKey: 'class_berserker_desc', color: '#e74c3c' },
];

export default function CharacterCreateScreen({ navigation }: Props) {
  const [name, setName] = useState('');
  const [selectedClass, setSelectedClass] = useState<Class | null>(null);
  const [creating, setCreating] = useState(false);
  const setCharacter = useGameStore((s) => s.setCharacter);
  const language = useGameStore((s) => s.language);

  async function handleCreate() {
    console.log('handleCreate called', { name, selectedClass, creating });
    if (!name.trim() || !selectedClass) {
      console.log('handleCreate blocked: name or class missing');
      return;
    }
    setCreating(true);
    try {
      console.log('Calling RPC create_character...');
      const result = await rpc<{ character: Character }>('create_character', {
        name: name.trim(),
        class: selectedClass,
      });
      console.log('RPC success:', result.character?.id);
      setCharacter(result.character);
      navigation.replace('Game', { characterId: result.character.id });
    } catch (err) {
      console.error('Failed to create character:', err);
      setCreating(false);
    }
  }

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <Text style={styles.label}>{t('character_name', language)}</Text>
      <TextInput
        style={styles.input}
        value={name}
        onChangeText={setName}
        placeholder={t('enter_name', language)}
        placeholderTextColor="#666"
        maxLength={20}
      />

      <Text style={styles.label}>{t('choose_class', language)}</Text>
      {classes.map((cls) => (
        <Pressable
          key={cls.id}
          style={[
            styles.classCard,
            selectedClass === cls.id && { borderColor: cls.color, borderWidth: 2 },
          ]}
          onPress={() => setSelectedClass(cls.id)}
        >
          <Text style={[styles.className, { color: cls.color }]}>{t(cls.nameKey, language)}</Text>
          <Text style={styles.classDesc}>{t(cls.descKey, language)}</Text>
        </Pressable>
      ))}

      <Pressable
        style={[styles.createButton, (!name.trim() || !selectedClass) && styles.buttonDisabled]}
        onPress={handleCreate}
        disabled={!name.trim() || !selectedClass || creating}
      >
        <Text style={styles.createButtonText}>
          {creating ? t('creating', language) : t('begin_adventure', language)}
        </Text>
      </Pressable>
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
  },
  label: {
    color: '#e0d68a',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 8,
    marginTop: 16,
  },
  input: {
    backgroundColor: '#1a1a2e',
    color: '#fff',
    padding: 16,
    borderRadius: 8,
    fontSize: 16,
  },
  classCard: {
    backgroundColor: '#1a1a2e',
    padding: 16,
    borderRadius: 8,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  className: {
    fontSize: 20,
    fontWeight: 'bold',
  },
  classDesc: {
    color: '#a0a0b0',
    fontSize: 14,
    marginTop: 4,
  },
  createButton: {
    backgroundColor: '#e0d68a',
    paddingVertical: 16,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 24,
  },
  buttonDisabled: {
    opacity: 0.4,
  },
  createButtonText: {
    color: '#1a1a2e',
    fontSize: 18,
    fontWeight: 'bold',
  },
});
