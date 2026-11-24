using System.Globalization;
using System.Text;

namespace Tabanco.Core;

/// <summary>Draws a bill as it would look chalked on the bar.</summary>
public static class Chalk
{
    /// <summary>
    /// Cents the Spanish way, "16,50 €". Done by hand rather than with the
    /// es-ES culture, because that culture comes from the machine's ICU data
    /// and a container running in globalization-invariant mode has no es-ES
    /// at all. By hand, it matches the Go version byte for byte everywhere.
    /// </summary>
    public static string Euros(int cents) =>
        string.Create(CultureInfo.InvariantCulture, $"{cents / 100},{cents % 100:00} €");

    /// <summary>n as chalk marks, in gates of five.</summary>
    internal static string Tally(int n)
    {
        var gates = new List<string>();
        for (; n >= 5; n -= 5)
        {
            gates.Add("||||/");
        }
        if (n > 0)
        {
            gates.Add(new string('|', n));
        }
        return string.Join(' ', gates);
    }

    public static string Draw(Bill bill)
    {
        var s = new StringBuilder();
        s.Append(bill.Name).Append('\n');
        if (bill.Lines.Count == 0)
        {
            s.Append("  (nothing chalked yet)\n");
            return s.ToString();
        }
        foreach (var l in bill.Lines)
        {
            s.Append(CultureInfo.InvariantCulture, $"  {l.Name,-14} {Tally(l.Count),-16} {l.Count,3} {Euros(l.Cents),10}\n");
        }
        s.Append("  ").Append('-', 46).Append('\n');
        s.Append(CultureInfo.InvariantCulture, $"  {"total",-35} {Euros(bill.Total),10}\n");
        return s.ToString();
    }
}
