import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import FlowchartPuzzle from '../src/lib/components/FlowchartPuzzle.svelte';

const solution = JSON.stringify([
  { shape: 'terminator', label: 'Mulai' },
  { shape: 'process', label: 'a = 1' },
  { shape: 'terminator', label: 'Selesai' },
]);

const setup = (
  props: { disabled?: boolean; distractors?: { shape: 'io'; label: string }[] } = {},
) => {
  const onresult = vi.fn();
  render(FlowchartPuzzle, { code: 'a = 1', solutionJson: solution, onresult, ...props });
  return { onresult };
};

// Ketuk kartu di bank lalu ketuk slot kosong pertama, seperti pemakaian di layar sentuh.
const place = async (label: string) => {
  await fireEvent.click(screen.getByRole('button', { name: label }));
  await fireEvent.click(screen.getAllByRole('button', { name: 'Slot kosong' })[0]);
};

describe('FlowchartPuzzle', () => {
  it('semua kartu dan pengecoh muncul di bank, tombol periksa mati', () => {
    setup({ distractors: [{ shape: 'io', label: 'Pengecoh' }] });
    for (const label of ['Mulai', 'a = 1', 'Selesai', 'Pengecoh']) {
      expect(screen.getByRole('button', { name: label })).toBeInTheDocument();
    }
    expect(screen.getAllByRole('button', { name: 'Slot kosong' })).toHaveLength(3);
    expect(screen.getByRole('button', { name: 'Periksa Susunan' })).toBeDisabled();
  });

  it('susunan benar dilaporkan benar', async () => {
    const { onresult } = setup();
    await place('Mulai');
    await place('a = 1');
    await place('Selesai');
    await fireEvent.click(screen.getByRole('button', { name: 'Periksa Susunan' }));
    expect(onresult).toHaveBeenLastCalledWith(true);
    expect(screen.getByRole('status')).toHaveTextContent('Susunan flowchart tepat');
  });

  it('susunan tertukar dilaporkan salah', async () => {
    const { onresult } = setup();
    await place('Selesai');
    await place('a = 1');
    await place('Mulai');
    await fireEvent.click(screen.getByRole('button', { name: 'Periksa Susunan' }));
    expect(onresult).toHaveBeenLastCalledWith(false);
    expect(screen.getByRole('status')).toHaveTextContent('Belum tepat');
  });

  it('mengetuk slot terisi mengembalikan kartu ke bank', async () => {
    setup();
    await place('Mulai');
    await fireEvent.click(screen.getByRole('button', { name: 'Lepas Mulai' }));
    expect(screen.getAllByRole('button', { name: 'Slot kosong' })).toHaveLength(3);
    expect(screen.getByRole('button', { name: 'Mulai' })).toBeInTheDocument();
  });

  it('Acak Ulang mengosongkan susunan dan melaporkan salah', async () => {
    const { onresult } = setup();
    await place('Mulai');
    await fireEvent.click(screen.getByRole('button', { name: 'Acak Ulang' }));
    expect(onresult).toHaveBeenLastCalledWith(false);
    expect(screen.getAllByRole('button', { name: 'Slot kosong' })).toHaveLength(3);
  });

  it('mode disabled (lesson sudah selesai) tidak menerima kartu', async () => {
    setup({ disabled: true });
    expect(screen.getByRole('button', { name: 'Mulai' })).toBeDisabled();
    await fireEvent.click(screen.getAllByRole('button', { name: 'Slot kosong' })[0]);
    expect(screen.getAllByRole('button', { name: 'Slot kosong' })).toHaveLength(3);
  });
});
