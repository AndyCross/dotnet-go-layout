using Tabanco.Core;
using Tabanco.Core.Storage;

const string Usage = """
    usage: chalk [--file chalk.json] <command>

      menu                    what the bar sells
      open <name>             open a tab
      order <name> <item>...  chalk up a round
      penultima <name>        one more of the usual
      show [name]             one tab, or every open tab
      settle <name>           price it up and wipe it off

    """;

var file = "chalk.json";
var rest = args.ToList();
if (rest.Count >= 2 && rest[0] is "--file" or "-file")
{
    file = rest[1];
    rest.RemoveRange(0, 2);
}
if (rest.Count == 0)
{
    Console.Error.Write(Usage);
    return 2;
}

try
{
    Run(new FileTabStore(file), Menu.Default, rest[0], rest[1..]);
    return 0;
}
catch (Exception e) when (e is TabRuleException or ChalkException)
{
    Console.Error.WriteLine($"chalk: {e.Message}");
    return 1;
}

static void Run(ITabStore store, Menu menu, string command, List<string> args)
{
    string NeedName() => args.Count > 0 ? args[0] : throw new ChalkException($"{command} needs a name");
    Tab Load(string name) => store.TryGet(name, out var tab) ? tab : throw new ChalkException("no tab open under that name");

    switch (command)
    {
        case "menu":
            foreach (var item in menu.Items)
            {
                Console.WriteLine($"  {item.Id,-14} {item.Name,-14} {item.Kind,-6} {Chalk.Euros(item.Price),8}");
            }
            break;

        case "open":
            var name = NeedName();
            if (store.TryGet(name, out _))
            {
                throw new ChalkException($"{name} already has a tab open");
            }
            store.Put(new Tab(name));
            break;

        case "order":
            var ordering = Load(NeedName());
            ordering.Order(menu, args[1..]);
            Console.Write(Chalk.Draw(ordering.Price(menu)));
            store.Put(ordering);
            break;

        case "penultima":
            var regular = Load(NeedName());
            var poured = regular.Penultima(menu);
            Console.WriteLine($"la penúltima: {poured.Name}");
            Console.Write(Chalk.Draw(regular.Price(menu)));
            store.Put(regular);
            break;

        case "show" when args.Count > 0:
            Console.Write(Chalk.Draw(Load(args[0]).Price(menu)));
            break;

        case "show":
            var open = store.List();
            if (open.Count == 0)
            {
                Console.WriteLine("the bar is empty");
            }
            foreach (var tab in open)
            {
                Console.Write(Chalk.Draw(tab.Price(menu)));
            }
            break;

        case "settle":
            var settling = Load(NeedName());
            Console.Write(Chalk.Draw(settling.Price(menu)));
            Console.WriteLine("settled. ¡Salud!");
            store.Delete(settling.Name);
            break;

        default:
            throw new ChalkException($"no command called {command}");
    }
}

internal sealed class ChalkException(string message) : Exception(message);
