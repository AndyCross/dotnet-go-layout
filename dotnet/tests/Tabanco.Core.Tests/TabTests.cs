namespace Tabanco.Core.Tests;

public class TabTests
{
    private static readonly Menu Menu = Menu.Default;

    [Fact]
    public void OrderAndPrice()
    {
        var tab = new Tab("Mari");
        tab.Order(Menu, "fino", "fino", "chicharrones");
        var bill = tab.Price(Menu);
        Assert.Equal(800, bill.Total);
        Assert.Equal(2, bill.Lines.Count);
        Assert.Equal(new Line("fino", "Fino", 2, 400), bill.Lines[0]);
    }

    [Fact]
    public void OrderWithUnknownItemMarksNothing()
    {
        var tab = new Tab("Mari");
        var e = Assert.Throws<UnknownItemException>(() => tab.Order(Menu, "fino", "cerveza"));
        Assert.Equal("cerveza", e.Id);
        Assert.Empty(tab.Marks);
    }

    [Fact]
    public void Penultima()
    {
        var tab = new Tab("Paco");
        Assert.Throws<NothingToRepeatException>(() => tab.Penultima(Menu));
        // Tapas never count, and a tie goes to the drink higher on the board.
        tab.Order(Menu, "chicharrones", "chicharrones", "chicharrones", "oloroso", "manzanilla");
        Assert.Equal("manzanilla", tab.Penultima(Menu).Id);
    }

    [Theory]
    [InlineData(0, "")]
    [InlineData(1, "|")]
    [InlineData(4, "||||")]
    [InlineData(5, "||||/")]
    [InlineData(7, "||||/ ||")]
    [InlineData(10, "||||/ ||||/")]
    public void Tally(int n, string want) => Assert.Equal(want, Chalk.Tally(n));

    [Fact]
    public void Draw()
    {
        var tab = new Tab("Mari");
        tab.Order(Menu, "fino", "fino", "fino", "fino", "fino", "fino", "fino", "px", "chicharrones");
        const string Want = """
            Mari
              Fino           ||||/ ||           7    14,00 €
              Pedro Ximénez  |                  1     2,50 €
              Chicharrones   |                  1     4,00 €
              ----------------------------------------------
              total                                  20,50 €

            """;
        Assert.Equal(Want, Chalk.Draw(tab.Price(Menu)));
    }
}
