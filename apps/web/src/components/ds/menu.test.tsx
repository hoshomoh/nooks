import { beforeAll, describe, expect, it, vi, type Mock } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import { readyForEnglish } from "@/test/i18n"
import { Menu, MenuItem } from "./menu"

beforeAll(readyForEnglish)

/** What each entry does, so a test can say which one a key reached. */
interface Pressed {
  pin: Mock<() => void>
  archive: Mock<() => void>
  rename: Mock<() => void>
}

/** show opens a menu holding the keys a List's menu holds. */
async function show(): Promise<Pressed> {
  const pressed: Pressed = {
    pin: vi.fn<() => void>(),
    archive: vi.fn<() => void>(),
    rename: vi.fn<() => void>(),
  }

  render(
    <Menu trigger={<button type="button">More</button>}>
      <MenuItem onSelect={pressed.rename}>Rename</MenuItem>
      <MenuItem shortcut="⌘D" onSelect={pressed.pin}>
        Pin
      </MenuItem>
      <MenuItem shortcut="⌘⇧A" onSelect={pressed.archive}>
        Archive
      </MenuItem>
    </Menu>,
  )

  await userEvent.click(screen.getByRole("button", { name: "More" }))
  await screen.findByRole("menu")
  return pressed
}

/*
A keycap a menu prints is a key the menu answers.

Every one of these was drawn and none of them did anything: the keycap was a label and
nothing listened for it. What stops that coming back is pressing the key.
*/
describe("the keys a menu prints", () => {
  it("does what the entry that printed it does", async () => {
    const pressed = await show()

    await userEvent.keyboard("{Meta>}d{/Meta}")

    expect(pressed.pin).toHaveBeenCalledOnce()
    expect(pressed.archive).not.toHaveBeenCalled()
  })

  // One keycap, two keyboards: ⌘ is what a Mac calls the key Ctrl is everywhere else.
  it("answers control where there is no command key", async () => {
    const pressed = await show()

    await userEvent.keyboard("{Control>}d{/Control}")

    expect(pressed.pin).toHaveBeenCalledOnce()
  })

  it("wants every modifier the keycap drew", async () => {
    const pressed = await show()

    await userEvent.keyboard("{Meta>}{Shift>}a{/Shift}{/Meta}")

    expect(pressed.archive).toHaveBeenCalledOnce()
    expect(pressed.pin).not.toHaveBeenCalled()
  })

  it("leaves a key alone when no entry printed it", async () => {
    const pressed = await show()

    await userEvent.keyboard("{Meta>}j{/Meta}")

    expect(pressed.pin).not.toHaveBeenCalled()
    expect(pressed.archive).not.toHaveBeenCalled()
    expect(pressed.rename).not.toHaveBeenCalled()
  })

  // An entry with no keycap is reached by choosing it, and by nothing else.
  it("does not invent a key for an entry that printed none", async () => {
    const pressed = await show()

    await userEvent.keyboard("{F2}")

    expect(pressed.rename).not.toHaveBeenCalled()
  })
})
