import React, { useEffect, useState } from 'react';
import { View, StyleSheet } from 'react-native';
import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { RootStackParamList } from '../../App';
import { createMatch } from '../api/socket';
import { useGameStore } from '../state/gameStore';
import NarrativeView from '../components/NarrativeView';
import ActionBar from '../components/ActionBar';
import CharacterSheet from '../components/CharacterSheet';
import MapView from '../components/MapView';
import CombatHUD from '../components/CombatHUD';
import ChatPanel from '../components/ChatPanel';

type Props = NativeStackScreenProps<RootStackParamList, 'Game'>;

export default function GameScreen({ route }: Props) {
  const { characterId } = route.params;
  const [showCharSheet, setShowCharSheet] = useState(false);
  const [showMap, setShowMap] = useState(false);
  const [showChat, setShowChat] = useState(false);
  const phase = useGameStore((s) => s.phase);

  useEffect(() => {
    async function start() {
      try {
        useGameStore.getState().setLoading(true);
        await createMatch(characterId);
      } catch (err) {
        console.error('Failed to create match:', err);
        useGameStore.getState().setLoading(false);
      }
    }
    start();
  }, [characterId]);

  if (showMap) {
    return <MapView onClose={() => setShowMap(false)} />;
  }

  const biomeColor = getBiomeColor(phase);

  return (
    <View style={[styles.container, { backgroundColor: biomeColor }]}>
      <CharacterSheet visible={showCharSheet} onClose={() => setShowCharSheet(false)} />
      <CombatHUD />
      <NarrativeView />
      {showChat && <ChatPanel onClose={() => setShowChat(false)} />}
      <ActionBar
        onCharacterSheet={() => setShowCharSheet(!showCharSheet)}
        onMap={() => setShowMap(true)}
        onChat={() => setShowChat(!showChat)}
      />
    </View>
  );
}

function getBiomeColor(phase: string): string {
  switch (phase) {
    case 'in_combat':
      return '#1a0a0a';
    case 'in_dialogue':
      return '#1a1a2e';
    default:
      return '#0d1b0d';
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingTop: 50,
    paddingBottom: 20,
  },
});
